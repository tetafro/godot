// Package godot checks if comments contain a period at the end of the last
// sentence if needed.
package godot

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/ioutil"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// CAUTION: Line and column indexes are 1-based.

// NOTE: Errors `invalid line number inside comment...` should never happen.
// Their goal is to prevent panic, if there's a bug with array indexes.

const (
	// noPeriodMessage is an error message to return.
	noPeriodMessage = "Comment should end in a period"
	// topLevelColumn is just the most left column of the file.
	topLevelColumn = 1
)

// Scope sets which comments should be checked.
type Scope string

// List of available check scopes.
const (
	// DeclScope is for top level declaration comments.
	DeclScope Scope = "decl"
	// TopLevelScope is for all top level comments.
	TopLevelScope Scope = "top"
	// AllScope is for all comments.
	AllScope Scope = "all"
)

// Settings contains linter settings.
type Settings struct {
	Scope Scope
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

// comment is an internal representation of AST comment entity with rendered
// lines attached. The latter is used for creating a full replacement for
// the line with issues.
type comment struct {
	ast   *ast.CommentGroup
	lines []string
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
func Run(file *ast.File, fset *token.FileSet, settings Settings) ([]Issue, error) {
	comments, err := getComments(file, fset, settings.Scope)
	if err != nil {
		return nil, fmt.Errorf("get comments: %v", err)
	}

	issues, err := checkComments(fset, comments)
	if err != nil {
		return nil, fmt.Errorf("check comments: %v", err)
	}

	sortIssues(issues)
	return issues, nil
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

	issues, err := Run(file, fset, settings)
	if err != nil {
		return nil, fmt.Errorf("run linter: %v", err)
	}

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

// getComments extracts comments from a file.
func getComments(file *ast.File, fset *token.FileSet, scope Scope) ([]comment, error) {
	var comments []comment

	// Render AST representation to a string
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("render file: %v", err)
	}
	lines := strings.Split(buf.String(), "\n")

	// All comments
	if scope == AllScope {
		cc, err := getAllComments(file, fset, lines)
		if err != nil {
			return nil, fmt.Errorf("get all comments: %v", err)
		}
		return append(comments, cc...), nil
	}

	// Comments from the inside of top level blocks
	cc, err := getBlockComments(file, fset, lines)
	if err != nil {
		return nil, fmt.Errorf("get block comments: %v", err)
	}
	comments = append(comments, cc...)

	// All top level comments
	if scope == TopLevelScope {
		cc, err := getTopLevelComments(file, fset, lines)
		if err != nil {
			return nil, fmt.Errorf("get top level comments: %v", err)
		}
		return append(comments, cc...), nil
	}

	// Top level declaration comments
	cc, err = getDeclarationComments(file, fset, lines)
	if err != nil {
		return nil, fmt.Errorf("get declaration comments: %v", err)
	}
	comments = append(comments, cc...)

	return comments, nil
}

// getBlockComments gets comments from the inside of top level
// blocks: var (...), const (...).
func getBlockComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment
	for _, decl := range file.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		// No parenthesis == no block
		if d.Lparen == 0 {
			continue
		}
		for _, c := range file.Comments {
			// Skip comments outside this block
			if d.Lparen > c.Pos() || c.Pos() > d.Rparen {
				continue
			}
			// Skip comments that are not top-level for this block
			if fset.Position(c.Pos()).Column != topLevelColumn+1 {
				continue
			}
			firstLine := fset.Position(c.Pos()).Line
			lastLine := fset.Position(c.End()).Line
			if lastLine >= len(lines) {
				return nil, fmt.Errorf(
					"invalid line number inside comment: %s:%d",
					fset.Position(c.Pos()).Filename,
					fset.Position(c.Pos()).Line,
				)
			}
			comments = append(comments, comment{
				ast:   c,
				lines: lines[firstLine-1 : lastLine],
			})
		}
	}
	return comments, nil
}

// getTopLevelComments gets all top level comments.
func getTopLevelComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment // nolint: prealloc
	for _, c := range file.Comments {
		if fset.Position(c.Pos()).Column != topLevelColumn {
			continue
		}
		firstLine := fset.Position(c.Pos()).Line
		lastLine := fset.Position(c.End()).Line
		if lastLine >= len(lines) {
			return nil, fmt.Errorf(
				"invalid line number inside comment: %s:%d",
				fset.Position(c.Pos()).Filename,
				fset.Position(c.Pos()).Line,
			)
		}
		comments = append(comments, comment{
			ast:   c,
			lines: lines[firstLine-1 : lastLine],
		})
	}
	return comments, nil
}

// getDeclarationComments gets top level declaration comments.
func getDeclarationComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Doc != nil {
				firstLine := fset.Position(d.Doc.Pos()).Line
				lastLine := fset.Position(d.Doc.End()).Line
				if lastLine >= len(lines) {
					return nil, fmt.Errorf(
						"invalid line number inside comment: %s:%d",
						fset.Position(d.Doc.Pos()).Filename,
						fset.Position(d.Doc.Pos()).Line,
					)
				}
				comments = append(comments, comment{
					ast:   d.Doc,
					lines: lines[firstLine-1 : lastLine],
				})
			}
		case *ast.FuncDecl:
			if d.Doc != nil {
				firstLine := fset.Position(d.Doc.Pos()).Line
				lastLine := fset.Position(d.Doc.End()).Line
				if lastLine >= len(lines) {
					return nil, fmt.Errorf(
						"invalid line number inside comment: %s:%d",
						fset.Position(d.Doc.Pos()).Filename,
						fset.Position(d.Doc.Pos()).Line,
					)
				}
				comments = append(comments, comment{
					ast:   d.Doc,
					lines: lines[firstLine-1 : lastLine],
				})
			}
		}
	}
	return comments, nil
}

// getAllComments gets every single comment from the file.
func getAllComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment //nolint: prealloc
	for _, c := range file.Comments {
		firstLine := fset.Position(c.Pos()).Line
		lastLine := fset.Position(c.End()).Line
		if lastLine >= len(lines) {
			return nil, fmt.Errorf(
				"invalid line number inside comment: %s:%d",
				fset.Position(c.Pos()).Filename,
				fset.Position(c.Pos()).Line,
			)
		}
		comments = append(comments, comment{
			ast:   c,
			lines: lines[firstLine-1 : lastLine],
		})
	}
	return comments, nil
}

// checkComments checks that every comment ends with a period.
func checkComments(fset *token.FileSet, comments []comment) ([]Issue, error) {
	var issues []Issue // nolint: prealloc
	for _, c := range comments {
		if c.ast == nil || len(c.ast.List) == 0 {
			continue
		}

		// Save global line number and indent
		start := fset.Position(c.ast.List[0].Slash)

		text := getText(c.ast)
		pos, ok := checkText(text)
		if ok {
			continue
		}

		iss := Issue{
			Pos: token.Position{
				Filename: start.Filename,
				Offset:   start.Offset,
				Line:     pos.line + start.Line - 1,
				Column:   pos.column + start.Column - 1,
			},
			Message: noPeriodMessage,
		}

		// Make a replacement. Use `pos.line` to get an original line from
		// attached lines. Use `iss.Pos.Column` because it's a position in
		// the original line.
		if pos.line-1 >= len(c.lines) {
			return nil, fmt.Errorf(
				"invalid line number inside comment: %s:%d",
				iss.Pos.Filename, iss.Pos.Line,
			)
		}
		original := []rune(c.lines[pos.line-1])
		if iss.Pos.Column-1 > len(original) {
			return nil, fmt.Errorf(
				"invalid column number inside comment: %s:%d:%d",
				iss.Pos.Filename, iss.Pos.Line, iss.Pos.Column,
			)
		}
		iss.Replacement = fmt.Sprintf("%s.%s",
			string(original[:iss.Pos.Column-1]),
			string(original[iss.Pos.Column-1:]))

		issues = append(issues, iss)
	}
	return issues, nil
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
		return ""
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
// of the issue if found.
// NOTE: Returned position is a position inside given text, not position in
// the original file.
func checkText(comment string) (pos position, ok bool) {
	isBlock := strings.HasPrefix(comment, "/*")

	// Check last non-empty line
	var found bool
	var line, prefix string
	lines := strings.Split(comment, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line = lines[i]

		// Trim //, /*, */ and save them
		prefix = ""
		if !isBlock {
			line = strings.TrimPrefix(line, "//")
			prefix = "//"
		}
		if isBlock && i == 0 {
			line = strings.TrimPrefix(line, "/*")
			prefix = "/*"
		}
		if isBlock && i == len(lines)-1 {
			line = strings.TrimSuffix(line, "*/")
		}

		line = strings.TrimRightFunc(line, unicode.IsSpace)
		if line == "" {
			continue
		}

		found = true
		pos.line = i + 1
		break
	}
	// All lines are empty
	if !found {
		return position{}, true
	}
	// Correct line
	if hasSuffix(line, lastChars) {
		return position{}, true
	}

	pos.column = len([]rune(prefix+line)) + 1
	return pos, false
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
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}
	return false
}
