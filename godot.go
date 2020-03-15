// Package godot checks if all top-level comments contain a period at the
// end of the last sentence if needed.
package godot

import (
	"go/ast"
	"go/token"
	"strings"
)

// Message contains a message of linting error.
type Message struct {
	Pos     token.Position
	Message string
}

// List of valid last characters.
var lastChars = []string{".", "?", "!"}

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
		last := group.List[len(group.List)-1].Text

		if !checkComment(last) {
			msgs = append(msgs, Message{
				Pos:     fset.Position(group.Pos()),
				Message: "Top level comment should end in a period",
			})
		}
	}
	return msgs
}

func checkComment(comment string) bool {
	// Check last line of "//"-comment
	if strings.HasPrefix(comment, "//") {
		comment = strings.TrimPrefix(comment, "//")
		return checkLastChar(comment)
	}

	// Check multiline "/*"-comment block
	lines := strings.Split(comment, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) == "*/" || lines[i] == "" {
			continue
		}
		comment = strings.TrimPrefix(lines[i], "/*")
		comment = strings.TrimSuffix(comment, "*/")
		return checkLastChar(comment)
	}

	return true
}

func checkLastChar(s string) bool {
	// Don't check comments starting with space indentation - they may
	// contain code examples, which shouldn't end with period
	if strings.HasPrefix(s, "  ") {
		return true
	}
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "nolint:") {
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
