// Package godot checks if all top-level comments contain a period at the
// end of the last sentence if needed.
package godot

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
)

// CAUTION: Line and column indexes start from 1.

const (
	// noPeriodMessage is an error message to return.
	noPeriodMessage = "Top level comment should end in a period"
	// topLevelColumn is just the most left column of the file.
	topLevelColumn = 1
)

// Settings contains linter settings.
type Settings struct {
	// Check all top-level comments, not only declarations
	CheckAll bool
}

// Issue contains a description of linting error and a recommended replacement.
type Issue struct {
	Pos         token.Position
	Message     string
	Replacement string
}

// position is a position inside a comment (might be multiline comment).
type position struct {
	line   int
	column int
}

var (
	// List of valid sentence ending.
	// NOTE: Sentence can be inside parenthesis, and therefore ends
	// with parenthesis.
	lastChars = []string{".", "?", "!", ".)", "?)", "!)"}

	// Special tags in comments like "// nolint:", or "// +k8s:".
	tags = regexp.MustCompile(`^\+?[a-z0-9]+:`)

	// Special hashtags in comments like "// #nosec".
	hashtags = regexp.MustCompile(`^#[a-z]+($|\s)`)

	// URL at the end of the line.
	endURL = regexp.MustCompile(`[a-z]+://[^\s]+$`)
)

// Run runs this linter on the provided code.
func Run(file *ast.File, fset *token.FileSet, settings Settings) []Issue {
	comments := getComments(file, fset, settings.CheckAll)
	issues := checkComments(fset, comments)
	sortIssues(issues)
	return issues
}

// Fix fixes all issues and returns new version of file content.
func Fix(path string, file *ast.File, fset *token.FileSet, settings Settings) ([]byte, error) {
	// Read file
	content, err := ioutil.ReadFile(path) // nolint: gosec
	if err != nil {
		return nil, fmt.Errorf("read file: %v", err)
	}
	if len(content) == 0 {
		return nil, nil
	}

	issues := Run(file, fset, settings)

	// slice -> map
	m := map[int]Issue{}
	for _, iss := range issues {
		m[iss.Pos.Line] = iss
	}

	// Replace lines from issues
	fixed := make([]byte, 0, len(content))
	for i, line := range strings.Split(string(content), "\n") {
		newline := line
		if iss, ok := m[i+1]; ok {
			newline = iss.Replacement
		}
		fixed = append(fixed, []byte(newline+"\n")...)
	}
	fixed = fixed[:len(fixed)-1] // trim last "\n"

	return fixed, nil
}

// Replace rewrites original file with it's fixed version.
func Replace(path string, file *ast.File, fset *token.FileSet, settings Settings) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("check file: %v", err)
	}
	mode := info.Mode()

	fixed, err := Fix(path, file, fset, settings)
	if err != nil {
		return fmt.Errorf("fix issues: %v", err)
	}

	if err := ioutil.WriteFile(path, fixed, mode); err != nil {
		return fmt.Errorf("write file: %v", err)
	}
	return nil
}

// sortIssues sorts by filename, line and column.
func sortIssues(iss []Issue) {
	sort.Slice(iss, func(i, j int) bool {
		if iss[i].Pos.Filename != iss[j].Pos.Filename {
			return iss[i].Pos.Filename < iss[j].Pos.Filename
		}
		if iss[i].Pos.Line != iss[j].Pos.Line {
			return iss[i].Pos.Line < iss[j].Pos.Line
		}
		return iss[i].Pos.Column < iss[j].Pos.Column
	})
}

// getComments extracts comments from a file. If `all` is set, all top-level
// comments are extracted, otherwise - only top-level declaration comments.
func getComments(file *ast.File, fset *token.FileSet, all bool) []*ast.CommentGroup {
	var comments []*ast.CommentGroup

	// Get comments from top level blocks: var (...), const (...)
	for _, decl := range file.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		// No parenthesis == no block
		if d.Lparen == 0 {
			continue
		}
		for _, group := range file.Comments {
			// Skip comments outside this block
			if d.Lparen > group.Pos() || group.Pos() > d.Rparen {
				continue
			}
			// Skip comments that are not top-level for this block
			if fset.Position(group.Pos()).Column != topLevelColumn+1 {
				continue
			}
			comments = append(comments, group)
		}
	}

	// Get all top level comments
	if all {
		for _, comment := range file.Comments {
			if fset.Position(comment.Pos()).Column != topLevelColumn {
				continue
			}
			comments = append(comments, comment)
		}
		return comments
	}

	// Get top level declaration comments
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Doc != nil {
			comments = append(comments, d.Doc)
			}
		case *ast.FuncDecl:
			if d.Doc != nil {
			comments = append(comments, d.Doc)
		}
	}
	}
	return comments
}

// checkComments checks that every comment ends with a period.
func checkComments(fset *token.FileSet, comments []*ast.CommentGroup) []Issue {
	var issues []Issue // nolint: prealloc
	for _, comment := range comments {
		if comment == nil || len(comment.List) == 0 {
			continue
		}

		start := fset.Position(comment.List[0].Slash)
		indent := strings.Repeat("\t", start.Column-1)

		text := getText(comment)
		pos, rep, ok := checkText(text)
		if ok {
			continue
		}

		iss := Issue{
			Pos: token.Position{
				Filename: start.Filename,
				Offset:   start.Offset,
				Line:     start.Line + pos.line - 1,
				Column:   start.Column + pos.column - 1,
			},
			Message:     noPeriodMessage,
			Replacement: indent + rep,
		}

		issues = append(issues, iss)
	}
	return issues
}

// getText extracts text from comment. If comment is a special block
// (e.g., CGO code), a block of empty lines is returned. If comment contains
// special lines (e.g., tags or indented code examples), they are replaced
// with a period, it's a hack to not force setting a period in comments
// before special lines. The result can be multiline.
func getText(comment *ast.CommentGroup) (s string) {
	if len(comment.List) == 1 &&
		strings.HasPrefix(comment.List[0].Text, "/*") &&
		isSpecialBlock(comment.List[0].Text) {
		return strings.Repeat("\n", len(comment.List[0].Text)-1)
	}

	for _, c := range comment.List {
		isMultiline := strings.HasPrefix(c.Text, "/*")
		for _, line := range strings.Split(c.Text, "\n") {
			if isSpecialLine(line) {
				if isMultiline {
					line = "."
				} else {
					line = "// ."
				}
			}
			s += line + "\n"
		}
	}
	if len(s) == 0 {
		return ""
	}
	return s[:len(s)-1] // trim last "\n"
}

// checkText checks extracted text from comment structure, and returns position
// of the an issue if found and a replacement for it.
// NOTE: Returned position is a position inside given text, not position in
// the original file.
func checkText(comment string) (pos position, replacement string, ok bool) {
	isBlock := strings.HasPrefix(comment, "/*")

	// Check last non-empty line
	var found bool
	var line string
	var prefix, suffix string
	lines := strings.Split(comment, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line = lines[i]

		// Trim //, /*, */ and save them
		suffix, prefix = "", ""
		if !isBlock {
			line = strings.TrimPrefix(line, "//")
			prefix = "//"
		}
		if isBlock && i == 0 {
			line = strings.TrimPrefix(line, "/*")
			prefix = "/*"
		}
		if isBlock && i == len(lines)-1 {
			if strings.HasSuffix(line, " */") {
				line = strings.TrimSuffix(line, " */")
				suffix = " */"
			}
			if strings.HasSuffix(line, "*/") {
				line = strings.TrimSuffix(line, "*/")
				suffix = "*/"
			}
		}

		if strings.TrimSpace(line) == "" {
			continue
		}

		found = true
		pos.line = i + 1
		break
	}
	// All lines are empty
	if !found {
		return position{}, "", true
	}
	// Correct line
	if hasSuffix(strings.TrimSpace(line), lastChars) {
		return position{}, "", true
	}

	pos.column = len([]rune(prefix+line)) + 1
	replacement = prefix + line + "." + suffix
	return pos, replacement, false
}

// isSpecialBlock checks that given block of comment lines is special and
// shouldn't be checked as a regular sentence.
func isSpecialBlock(comment string) bool {
	// Skip cgo code blocks
	// TODO: Find a better way to detect cgo code
	if strings.HasPrefix(comment, "/*") && (strings.Contains(comment, "#include") ||
		strings.Contains(comment, "#define")) {
		return true
	}
	return false
}

// isSpecialBlock checks that given comment line is special and
// shouldn't be checked as a regular sentence.
func isSpecialLine(comment string) bool {
	// Skip cgo export tags: https://golang.org/cmd/cgo/#hdr-C_references_to_Go
	if strings.HasPrefix(comment, "//export ") {
		return true
	}

	comment = strings.TrimPrefix(comment, "//")
	comment = strings.TrimPrefix(comment, "/*")

	// Don't check comments starting with space indentation - they may
	// contain code examples, which shouldn't end with period
	if strings.HasPrefix(comment, "  ") ||
		strings.HasPrefix(comment, " \t") ||
		strings.HasPrefix(comment, "\t") {
		return true
	}

	// Skip tags and URLs
	comment = strings.TrimSpace(comment)
	if tags.MatchString(comment) ||
		hashtags.MatchString(comment) ||
		endURL.MatchString(comment) ||
		strings.HasPrefix(comment, "+build") {
		return true
	}

	return false
}

func hasSuffix(s string, suffixes []string) bool {
	if s == "" {
		return false
	}
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}
