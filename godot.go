// Package godot checks if all top-level comments contain a period at the
// end of the last sentence if needed.
package godot

import (
	"go/ast"
	"go/token"
	"regexp"
	"strings"
)

// Message contains a message of linting error.
type Message struct {
	Pos     token.Position
	Message string
}

var (
	// List of valid last characters.
	lastChars = []string{".", "?", "!"}

	// Special tags in comments like "nolint" or "build".
	tags = regexp.MustCompile("^[a-z]+:")

	// URL at the end of the line
	endURL = regexp.MustCompile(`[a-z]+://[^\s]+$`)
)

// Run runs this linter on the provided code.
func Run(file *ast.File, fset *token.FileSet) []Message {
	msgs := []Message{}
	for _, group := range file.Comments {
		if len(group.List) == 0 {
			continue
		}

		// Check only top-level comments
		if fset.Position(group.Pos()).Column > 1 {
			continue
		}

		// Get last element from comment group - it can be either
		// last (or single) line for "//"-comment, or multiline string
		// for "/*"-comment
		last := group.List[len(group.List)-1]

		if line, ok := checkComment(last.Text); !ok {
			pos := fset.Position(last.Slash)
			pos.Line += line
			msgs = append(msgs, Message{
				Pos:     pos,
				Message: "Top level comment should end in a period",
			})
		}
	}
	return msgs
}

func checkComment(comment string) (line int, ok bool) {
	// Check last line of "//"-comment
	if strings.HasPrefix(comment, "//") {
		comment = strings.TrimPrefix(comment, "//")
		return 0, checkLastChar(comment)
	}

	// Check multiline "/*"-comment block
	lines := strings.Split(comment, "\n")
	var i int
	for i = len(lines) - 1; i >= 0; i-- {
		if s := strings.TrimSpace(lines[i]); s == "*/" || s == "" {
			continue
		}
		break
	}
	comment = strings.TrimPrefix(lines[i], "/*")
	comment = strings.TrimSuffix(comment, "*/")
	return i, checkLastChar(comment)
}

func checkLastChar(s string) bool {
	// Don't check comments starting with space indentation - they may
	// contain code examples, which shouldn't end with period
	if strings.HasPrefix(s, "  ") || strings.HasPrefix(s, "\t") {
		return true
	}
	s = strings.TrimSpace(s)
	if tags.MatchString(s) || endURL.MatchString(s) || strings.HasPrefix(s, "+build") {
		return true
	}
	// Don't check empty lines
	if s == "" {
		return true
	}
	for _, ch := range lastChars {
		if string(s[len(s)-1]) == ch {
			return true
		}
	}
	return false
}
