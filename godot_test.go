package godot

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetText(t *testing.T) {
	testCases := []struct {
		name    string
		comment *ast.CommentGroup
		text    string
	}{
		{
			name: "regular text",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// Hello, world"},
			}},
			text: "// Hello, world",
		},
		{
			name: "regular text without indentation",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "//Hello, world"},
			}},
			text: "//Hello, world",
		},
		{
			name: "empty singleline comment",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "//"},
			}},
			text: "//",
		},
		{
			name: "empty multiline comment",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/**/"},
			}},
			text: "/**/",
		},
		{
			name: "regular text in multiline block",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/*\nHello, world\n*/"},
			}},
			text: "/*\nHello, world\n*/",
		},
		{
			name: "block of singleline comments with regular text",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// One"},
				{Text: "// Two"},
				{Text: "// Three"},
			}},
			text: "// One\n// Two\n// Three",
		},
		{
			name: "block of singleline comments with empty and special lines",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// One"},
				{Text: "//"},
				{Text: "//  \t  "},
				{Text: "// Two"},
				{Text: "// #nosec"},
				{Text: "// Three"},
				{Text: "// +k8s:deepcopy-gen=package"},
				{Text: "// +nolint: gosec"},
			}},
			text: "// One\n//\n// .\n// Two\n// .\n// Three\n// .\n// .",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if text := getText(tt.comment); text != tt.text {
				t.Fatalf("Wrong text\n  expected: %s\n       got: %s", tt.text, text)
			}
		})
	}
}

func TestCheckText(t *testing.T) {
	testCases := []struct {
		name        string
		comment     string
		ok          bool
		issue       position
		replacement string
	}{
		{
			name:    "sentence with period, singleline comment",
			comment: "// Hello, world.",
			ok:      true,
		},
		{
			name:    "with period, no indentation",
			comment: "//Hello, world.",
			ok:      true,
		},
		{
			name:    "with period, multiple singleline comments",
			comment: "// Hello,\n// world.",
			ok:      true,
		},
		{
			name:    "with period, multiline block",
			comment: "/*\nHello, world.\n*/",
			ok:      true,
		},
		{
			name:    "with period, multiline block, single line",
			comment: "/* Hello, world. */",
			ok:      true,
		},
		{
			name:        "no period, singleline comment",
			comment:     "// Hello, world",
			issue:       position{line: 1, column: 16},
			replacement: "// Hello, world.",
		},
		{
			name:        "no period, multiline block",
			comment:     "/*\nHello, world\n*/",
			issue:       position{line: 2, column: 13},
			replacement: "Hello, world.",
		},
		{
			name:        "no period, multiline block, single line",
			comment:     "/* Hello, world */",
			issue:       position{line: 1, column: 16},
			replacement: "/* Hello, world. */",
		},
		{
			name:        "no period, set of single line comments",
			comment:     "// Hello,\n// world",
			issue:       position{line: 2, column: 9},
			replacement: "// world.",
		},
		{
			name:    "question mark",
			comment: "// Hello, world?",
			ok:      true,
		},
		{
			name:    "exclamation mark",
			comment: "// Hello, world!",
			ok:      true,
		},
		{
			name:    "empty line",
			comment: "//",
			ok:      true,
		},
		{
			name:    "empty multiline block",
			comment: "/**/",
			ok:      true,
		},
		{
			name:    "multiple empty singleline comments",
			comment: "//\n//\n//\n//",
			ok:      true,
		},
		{
			name:    "only spaces",
			comment: "//   ",
			ok:      true,
		},
		{
			name:    "mixed spaces",
			comment: "// \t\t  ",
			ok:      true,
		},
		{
			name:    "mixed spaces, multiline block",
			comment: "/* \t\t \n\n\n  \n\t  */",
			ok:      true,
		},
		{
			name:    "cyrillic, with period",
			comment: "// Кириллица.",
			ok:      true,
		},
		{
			name:        "// cyrillic, without period",
			comment:     "// Кириллица",
			issue:       position{line: 1, column: 13},
			replacement: "// Кириллица.",
		},
		{
			name:    "parenthesis, with period",
			comment: "// Hello. (World.)",
			ok:      true,
		},
		{
			name:        "parenthesis, without period",
			comment:     "// Hello. (World)",
			issue:       position{line: 1, column: 18},
			replacement: "// Hello. (World).",
		},
		{
			name:    "single closing parenthesis with period",
			comment: "// ).",
			ok:      true,
		},
		{
			name:        "single closing parenthesis without period",
			comment:     "// )",
			issue:       position{line: 1, column: 5},
			replacement: "// ).",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, rep, ok := checkText(tt.comment)
			if ok != tt.ok {
				t.Fatalf("Wrong result\n  expected: %v\n       got: %v", tt.ok, ok)
			}
			if pos.line != tt.issue.line {
				t.Fatalf("Wrong line\n  expected: %d\n       got: %d", tt.issue.line, pos.line)
			}
			if pos.column != tt.issue.column {
				t.Fatalf("Wrong column\n  expected: %d\n       got: %d", tt.issue.column, pos.column)
			}
			if rep != tt.replacement {
				t.Fatalf("Wrong replacement\n  expected: %s\n       got: %s", tt.replacement, rep)
			}
		})
	}
}

func TestIsSpecialBlock(t *testing.T) {
	testCases := []struct {
		name      string
		comment   string
		isSpecial bool
	}{
		{
			name:      "regular text",
			comment:   "Hello, world.",
			isSpecial: false,
		},
		{
			name:      "comment",
			comment:   "// Hello, world",
			isSpecial: false,
		},
		{
			name:      "empty string",
			comment:   "",
			isSpecial: false,
		},
		{
			name:      "Multiline comment",
			comment:   "/*\nHello, world\n*/",
			isSpecial: false,
		},
		{
			name: "CGO block",
			comment: `/*
				#include <iostream>

				int main() {
					std::cout << "Hello World!";
					return 0;
				}
			*/`,
			isSpecial: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if isSpecialBlock(tt.comment) != tt.isSpecial {
				t.Fatalf("Wrong result")
			}
		})
	}
}

func TestIsSpecialLine(t *testing.T) {
	testCases := []struct {
		name      string
		comment   string
		isSpecial bool
	}{
		{
			name:      "regular text",
			comment:   "// Hello, world.",
			isSpecial: false,
		},
		{
			name:      "regular text without period",
			comment:   "// Hello, world",
			isSpecial: false,
		},
		{
			name:      "code example (two spaces indentation)",
			comment:   "//  x == y",
			isSpecial: true,
		},
		{
			name:      "code example (many spaces indentation)",
			comment:   "//  x == y",
			isSpecial: true,
		},
		{
			name:      "code example (single tab indentation)",
			comment:   "//\tx == y",
			isSpecial: true,
		},
		{
			name:      "code example (many tabs indentation)",
			comment:   "// \t\t\tx == y",
			isSpecial: true,
		},
		{
			name:      "code example (mixed indentation)",
			comment:   "//  \t  \tx == y",
			isSpecial: true,
		},
		{
			name:      "nolint tag",
			comment:   "// nolint: test",
			isSpecial: true,
		},
		{
			name:      "nolint tag without indentation",
			comment:   "//nolint: test",
			isSpecial: true,
		},
		{
			name:      "build tags",
			comment:   "// +build !linux",
			isSpecial: true,
		},
		{
			name:      "build tags without indentation",
			comment:   "//+build !linux",
			isSpecial: true,
		},
		{
			name:      "kubernetes tag",
			comment:   "// +k8s:deepcopy-gen=package",
			isSpecial: true,
		},
		{
			name:      "cgo exported function",
			comment:   "//export FuncName",
			isSpecial: true,
		},
		{
			name:      "cgo exported function with indentation (wrong format)",
			comment:   "// export FuncName",
			isSpecial: false,
		},
		{
			name:      "url at the end of line",
			comment:   "// Read more: http://example.com/",
			isSpecial: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if isSpecialLine(tt.comment) != tt.isSpecial {
				t.Fatalf("Wrong result")
			}
		})
	}
}

func TestHasSuffix(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		suffixes []string
		result   bool
	}{
		{
			name:     "has",
			text:     "hello, world.",
			suffixes: []string{",", "?", "."},
			result:   true,
		},
		{
			name:     "has not",
			text:     "hello, world.",
			suffixes: []string{",", "?", ":"},
			result:   false,
		},
		{
			name:     "has, long suffixes",
			text:     "hello, world.",
			suffixes: []string{"x.", "m.", "d."},
			result:   true,
		},
		{
			name:     "has not, long suffixes",
			text:     "hello, world.",
			suffixes: []string{"x.", "m.", "k."},
			result:   false,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if hasSuffix(tt.text, tt.suffixes) != tt.result {
				t.Fatalf("Wrong result")
			}
		})
	}
}

func TestRunIntegration(t *testing.T) {
	testCases := []struct {
		name     string
		fileIn   string
		checkAll bool
	}{
		{
			name:     "default check",
			fileIn:   filepath.Join("testdata", "default", "in", "main.go"),
			checkAll: false,
		},
		{
			name:     "check all",
			fileIn:   filepath.Join("testdata", "checkall", "in", "main.go"),
			checkAll: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			expected, err := readTestFile(tt.fileIn)
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.fileIn, err)
			}

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tt.fileIn, nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse file %s: %v", tt.fileIn, err)
			}

			issues := Run(file, fset, Settings{CheckAll: tt.checkAll})
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
	}
}

func TestFixIntegration(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		path := filepath.Join("testdata", "not-exists.go")
		_, err := Fix(path, nil, nil, Settings{})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join("testdata", "empty.go")
		fixed, err := Fix(path, nil, nil, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if fixed != nil {
			t.Fatalf("Unexpected result: %s", string(fixed))
		}
	})

	testCases := []struct {
		name     string
		fileIn   string
		fileOut  string
		checkAll bool
		errors   bool
	}{
		{
			name:     "default check",
			fileIn:   filepath.Join("testdata", "default", "in", "main.go"),
			fileOut:  filepath.Join("testdata", "default", "out", "main.go"),
			checkAll: false,
			errors:   false,
		},
		{
			name:     "check all",
			fileIn:   filepath.Join("testdata", "checkall", "in", "main.go"),
			fileOut:  filepath.Join("testdata", "checkall", "out", "main.go"),
			checkAll: true,
			errors:   false,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			expected, err := ioutil.ReadFile(tt.fileOut) // nolint: gosec
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.fileOut, err)
			}

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tt.fileIn, nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse file %s: %v", tt.fileIn, err)
			}

			fixed, err := Fix(tt.fileIn, file, fset, Settings{CheckAll: tt.checkAll})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			fixedLines := strings.Split(string(fixed), "\n")
			expectedLines := strings.Split(string(expected), "\n")
			if len(fixedLines) != len(expectedLines) {
				t.Fatalf("Invalid number of result lines\n  expected: %d\n       got: %d",
					len(expectedLines), len(fixedLines))
			}
			for i := range fixedLines {
				// NOTE: This is a fix for Windows, not sure why this is happening
				result := strings.TrimRight(fixedLines[i], "\r")
				exp := strings.TrimRight(expectedLines[i], "\r")
				if result != exp {
					t.Fatalf("Wrong line %d in fixed file\n  expected: %s\n       got: %s",
						i, exp, result)
				}
			}
		})
	}
}

func TestReplaceIntegration(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		path := filepath.Join("testdata", "not-exists.go")
		err := Replace(path, nil, nil, Settings{})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	testCases := []struct {
		name     string
		fileIn   string
		fileOut  string
		checkAll bool
	}{
		{
			name:     "default check",
			fileIn:   filepath.Join("testdata", "default", "in", "main.go"),
			fileOut:  filepath.Join("testdata", "default", "out", "main.go"),
			checkAll: false,
		},
		{
			name:     "check all",
			fileIn:   filepath.Join("testdata", "checkall", "in", "main.go"),
			fileOut:  filepath.Join("testdata", "checkall", "out", "main.go"),
			checkAll: true,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			info, err := os.Stat(tt.fileIn)
			if err != nil {
				t.Fatalf("Failed to check test file %s: %v", tt.fileIn, err)
			}
			mode := info.Mode()
			original, err := ioutil.ReadFile(tt.fileIn) // nolint: gosec
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.fileIn, err)
			}
			defer func() {
				ioutil.WriteFile(tt.fileIn, original, mode) // nolint: errcheck,gosec
			}()
			expected, err := ioutil.ReadFile(tt.fileOut) // nolint: gosec
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.fileOut, err)
			}

			fset := token.NewFileSet()
			file, err := parser.ParseFile(fset, tt.fileIn, nil, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse file %s: %v", tt.fileIn, err)
			}

			if err := Replace(tt.fileIn, file, fset, Settings{CheckAll: tt.checkAll}); err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			fixed, err := ioutil.ReadFile(tt.fileIn) // nolint: gosec
			if err != nil {
				t.Fatalf("Failed to read fixed file %s: %v", tt.fileIn, err)
			}

			fixedLines := strings.Split(string(fixed), "\n")
			expectedLines := strings.Split(string(expected), "\n")
			if len(fixedLines) != len(expectedLines) {
				t.Fatalf("Invalid number of result lines\n  expected: %d\n       got: %d",
					len(expectedLines), len(fixedLines))
			}
			for i := range fixedLines {
				// NOTE: This is a fix for Windows, not sure why this is happening
				result := strings.TrimRight(fixedLines[i], "\r")
				exp := strings.TrimRight(expectedLines[i], "\r")
				if result != exp {
					t.Fatalf("Wrong line %d in fixed file\n  expected: %s\n       got: %s",
						i, exp, result)
				}
			}
		})
	}
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
