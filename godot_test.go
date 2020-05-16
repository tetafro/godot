package godot

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckComment(t *testing.T) {
	testCases := []struct {
		name    string
		comment string
		ok      bool
		issue   issue
	}{
		// Single line comments
		{
			name:    "singleline comment: ok",
			comment: "// Hello, world.",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: no period",
			comment: "// Hello, world",
			ok:      false,
			issue:   issue{line: 0, column: 15},
		},
		{
			name:    "singleline comment: question mark",
			comment: "// Hello, world?",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: exclamation mark",
			comment: "// Hello, world!",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: code example without period",
			comment: "//  x == y",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: code example without period and long indentation",
			comment: "//       x == y",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: code example without period and tab indentation",
			comment: "//\tx == y",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: code example without period and mixed indentation",
			comment: "// \tx == y",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: empty line",
			comment: "//",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: empty line with spaces",
			comment: "//      ",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: without indentation and with period",
			comment: "//hello, world.",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: without indentation and without period",
			comment: "//hello, world",
			ok:      false,
			issue:   issue{line: 0, column: 14},
		},
		{
			name:    "singleline comment: nolint mark without period",
			comment: "// nolint: test",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: nolint mark without indentation without period",
			comment: "//nolint: test",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: build tags",
			comment: "// +build !linux",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: cgo exported function",
			comment: "//export FuncName",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "singleline comment: url at the end of line",
			comment: "// Read more: http://example.com/",
			ok:      true,
			issue:   issue{},
		},
		// Multiline comments
		{
			name:    "multiline comment: ok",
			comment: "/*\n" + "Hello, world.\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: no period",
			comment: "/*\n" + "Hello, world\n" + "*/",
			ok:      false,
			issue:   issue{line: 1, column: 12},
		},
		{
			name:    "multiline comment: question mark",
			comment: "/*\n" + "Hello, world?\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: exclamation mark",
			comment: "/*\n" + "Hello, world!\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: code example without period",
			comment: "/*\n" + "  x == y\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: code example without period and long indentation",
			comment: "/*\n" + "       x == y\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: code example without period and tab indentation",
			comment: "/*\n" + "\tx == y\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: empty line",
			comment: "/**/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: empty line inside",
			comment: "/*\n" + "\n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: empty line with spaces",
			comment: "/*\n" + "    \n" + "*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: closing with whitespaces",
			comment: "/*\n" + "    \n" + "   */",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: one-liner with period",
			comment: "/* Hello, world. */",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: one-liner with period and without indentation",
			comment: "/*Hello, world.*/",
			ok:      true,
			issue:   issue{},
		},
		{
			name:    "multiline comment: one-liner without period and without indentation",
			comment: "/*Hello, world*/",
			ok:      false,
			issue:   issue{line: 0, column: 14},
		},
		{
			name:    "multiline comment: long comment",
			comment: "/*\n" + "\n" + "   \n" + "Hello, world\n" + "\n" + "  \n" + "*/",
			ok:      false,
			issue:   issue{line: 3, column: 12},
		},
		{
			name:    "multiline comment: url at the end of line",
			comment: "/*\n" + "Read more: http://example.com/" + "*/",
			ok:      true,
			issue:   issue{},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			iss, ok := checkComment(tt.comment)
			if ok != tt.ok {
				t.Fatalf("Wrong result, expected %v, got %v", tt.ok, ok)
			}
			if iss.line != tt.issue.line {
				t.Fatalf("Wrong line, expected %d, got %d", tt.issue.line, iss.line)
			}
			if iss.column != tt.issue.column {
				t.Fatalf("Wrong column, expected %d, got %d", tt.issue.column, iss.column)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	t.Run("default check", func(t *testing.T) {
		var testFile = filepath.Join("testdata", "default", "example.go")
		expected, err := readTestFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read test file %s: %v", testFile, err)
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", testFile, err)
		}

		issues := Run(file, fset, Settings{CheckAll: false})
		if len(issues) != len(expected) {
			t.Fatalf("Invalid number of result issues\n  expected: %d\n       got: %d",
				len(expected), len(issues))
		}
		for i := range issues {
			if issues[i].Pos.Filename != expected[i].Pos.Filename ||
				issues[i].Pos.Line != expected[i].Pos.Line {
				t.Fatalf("Unexpected position\n  expected %s\n       got %s",
					expected[i].Pos, issues[i].Pos)
			}
		}
	})

	t.Run("check all", func(t *testing.T) {
		var testFile = filepath.Join("testdata", "checkall", "example.go")
		expected, err := readTestFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read test file %s: %v", testFile, err)
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", testFile, err)
		}

		issues := Run(file, fset, Settings{CheckAll: true})
		if len(issues) != len(expected) {
			t.Fatalf("Invalid number of result issues\n  expected: %d\n       got: %d",
				len(expected), len(issues))
		}
		for i := range issues {
			if issues[i].Pos.Filename != expected[i].Pos.Filename ||
				issues[i].Pos.Line != expected[i].Pos.Line {
				t.Fatalf("Unexpected position\n  expected %d:%d\n       got %d:%d",
					expected[i].Pos.Line, expected[i].Pos.Column,
					issues[i].Pos.Line, issues[i].Pos.Column,
				)
			}
		}
	})
}

// readTestFile reads comments from file. If comment contains "PASS",
// it should not be among issues. If comment contains "FAIL", it should
// be among error issues.
func readTestFile(file string) ([]Issue, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var issues []Issue
	for _, group := range f.Comments {
		if group == nil || len(group.List) == 0 {
			continue
		}
		for _, com := range group.List {
			// Check every line for multiline comments
			for i, line := range strings.Split(com.Text, "\n") {
				if strings.Contains(line, "PASS") {
					continue
				}
				if strings.Contains(line, "FAIL") {
					pos := fset.Position(com.Slash)
					pos.Line += i
					issues = append(issues, Issue{
						Pos:     pos,
						Message: noPeriodMessage,
					})
				}
			}
		}
	}
	return issues, nil
}
