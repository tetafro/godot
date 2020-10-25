package godot

import (
	"fmt"
	"go/token"
	"regexp"
	"strings"
	"unicode"
)

// Error messages.
const (
	noPeriodMessage  = "Comment should end in a period"
	noCapitalMessage = "Sentence should start with a capital letter"
)

var (
	// List of valid sentence ending.
	// A sentence can be inside parenthesis, and therefore ends with parenthesis.
	// A colon is a valid sentence ending, because it can be followed by a
	// code example which is not checked.
	lastChars = []string{".", "?", "!", ".)", "?)", "!)", ":"}

	// Special tags in comments like "// nolint:", or "// +k8s:".
	tags = regexp.MustCompile(`^\+?[a-z0-9]+:`)

	// Special hashtags in comments like "// #nosec".
	hashtags = regexp.MustCompile(`^#[a-z]+($|\s)`)

	// URL at the end of the line.
	endURL = regexp.MustCompile(`[a-z]+://[^\s]+$`)
)

// checkComments checks every comment accordings to the rules from
// `settings` argument.
func checkComments(fset *token.FileSet, comments []comment, settings Settings) ([]Issue, error) {
	var issues []Issue // nolint: prealloc
	for _, c := range comments {
		if c.ast == nil || len(c.ast.List) == 0 {
			continue
		}

		if settings.Period {
			iss, err := checkCommentForPeriod(fset, c)
			if err != nil {
				return nil, fmt.Errorf("check comment for period: %v", err)
			}
			if iss != nil {
				issues = append(issues, *iss)
			}
		}

		if settings.Capital {
			iss, err := checkCommentForCapital(fset, c)
			if err != nil {
				return nil, fmt.Errorf("check comment for capital: %v", err)
			}
			if len(iss) > 0 {
				issues = append(issues, iss...)
			}
		}
	}
	return issues, nil
}

// checkCommentForPeriod checks that the last sentense of the comment ends
// in a period.
func checkCommentForPeriod(fset *token.FileSet, c comment) (*Issue, error) {
	// Save global line number and indent
	start := fset.Position(c.ast.List[0].Slash)

	text := getText(c.ast)

	pos, ok := checkPeriod(text)
	if ok {
		return nil, nil
	}

	// Shift position by the length of comment's special symbols: /* or //
	isBlock := strings.HasPrefix(c.ast.List[0].Text, "/*")
	if (isBlock && pos.line == 1) || !isBlock {
		pos.column += 2
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

	return &iss, nil
}

// checkCommentForCapital checks that the each sentense of the comment starts with
// a capital letter.
// nolint: unparam
func checkCommentForCapital(fset *token.FileSet, c comment) ([]Issue, error) {
	// Save global line number and indent
	start := fset.Position(c.ast.List[0].Slash)

	text := getText(c.ast)

	pp := checkCapital(text, c.decl)
	if len(pp) == 0 {
		return nil, nil
	}

	issues := make([]Issue, len(pp))
	for i, pos := range pp {
		// Shift position by the length of comment's special symbols: /* or //
		isBlock := strings.HasPrefix(c.ast.List[0].Text, "/*")
		if (isBlock && pos.line == 1) || !isBlock {
			pos.column += 2
		}

		issues[i] = Issue{
			Pos: token.Position{
				Filename: start.Filename,
				Offset:   start.Offset,
				Line:     pos.line + start.Line - 1,
				Column:   pos.column + start.Column - 1,
			},
			Message: noCapitalMessage,
		}
	}

	return issues, nil
}

// checkPeriod checks that the last sentense of the text ends in a period.
// NOTE: Returned position is a position inside given text, not in the
// original file.
func checkPeriod(comment string) (pos position, ok bool) {
	// Check last non-empty line
	var found bool
	var line string
	lines := strings.Split(comment, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line = strings.TrimRightFunc(lines[i], unicode.IsSpace)
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

	pos.column = len([]rune(line)) + 1
	return pos, false
}

// checkCapital checks that the each sentense of the text starts with
// a capital letter.
// NOTE: First letter is not checked in declaration comments, because they
// can describe unexported functions, which start from small letter.
func checkCapital(comment string, skipFirst bool) (pp []position) {
	const empty, endChar, endOfSentence = 1, 2, 3

	pos := position{line: 1}
	state := endOfSentence
	if skipFirst {
		state = empty
	}
	for _, r := range comment {
		s := string(r)

		pos.column++
		if s == "\n" {
			pos.line++
			pos.column = 0
			if state == endChar {
				state = endOfSentence
			}
			continue
		}
		if s == "." || s == "!" || s == "?" {
			state = endChar
			continue
		}
		if s == ")" && state == endChar {
			continue
		}
		if s == " " {
			if state == endChar {
				state = endOfSentence
			}
			continue
		}
		if state == endOfSentence && unicode.IsLower(r) {
			pp = append(pp, position{line: pos.line, column: pos.column})
		}
		state = empty
	}
	return pp
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
