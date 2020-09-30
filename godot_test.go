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

func TestGetComments(t *testing.T) {
	testFile := filepath.Join("testdata", "get", "main.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse input file: %v", err)
	}

	testCases := []struct {
		name     string
		scope    Scope
		contains []string
	}{
		{
			name:     "scope: decl",
			scope:    DeclScope,
			contains: []string{"[DECL]"},
		},
		{
			name:     "scope: top",
			scope:    TopLevelScope,
			contains: []string{"[DECL]", "[TOP]"},
		},
		{
			name:     "scope: all",
			scope:    AllScope,
			contains: []string{"[DECL]", "[TOP]", "[ALL]"},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			comments, err := getComments(file, fset, tt.scope)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			var expected int
			for _, c := range comments {
				if strings.Contains(c.ast.Text(), "[NONE]") {
					continue
				}
				text := c.ast.Text()
				for _, s := range tt.contains {
					if strings.Contains(text, s) {
						expected++
						break
					}
				}
			}
			if len(comments) != expected {
				t.Fatalf(
					"Got wrong number of comments:\n  expected: %d\n       got: %d",
					expected, len(comments),
				)
			}
		})
	}
}

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
		{
			name: "cgo block",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/*\n" +
					"#include <stdio.h>\n" +
					"#include <stdlib.h>\n" +
					"void myprint(char* s) {\n" +
					"\tprintf(s);\n" +
					"}\n" +
					"*/"},
			}},
			text: "",
		},
		{
			name: "multiline block with a code example",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/*\n" +
					"Example:\n" +
					"\tn := rand.Int()\n" +
					"\tfmt.Println(n)\n" +
					"*/"},
			}},
			text: "/*\n" +
				"Example:\n" +
				".\n" +
				".\n" +
				"*/",
		},
		{
			name:    "empty comment group",
			comment: &ast.CommentGroup{List: []*ast.Comment{}},
			text:    "",
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
		name    string
		comment string
		ok      bool
		issue   position
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
			name:    "no period, singleline comment",
			comment: "// Hello, world",
			issue:   position{line: 1, column: 16},
		},
		{
			name:    "no period, multiline block",
			comment: "/*\nHello, world\n*/",
			issue:   position{line: 2, column: 13},
		},
		{
			name:    "no period, multiline block, single line",
			comment: "/* Hello, world */",
			issue:   position{line: 1, column: 16},
		},
		{
			name:    "no period, set of single line comments",
			comment: "// Hello,\n// world",
			issue:   position{line: 2, column: 9},
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
			name:    "// cyrillic, without period",
			comment: "// Кириллица",
			issue:   position{line: 1, column: 13},
		},
		{
			name:    "parenthesis, with period",
			comment: "// Hello. (World.)",
			ok:      true,
		},
		{
			name:    "parenthesis, without period",
			comment: "// Hello. (World)",
			issue:   position{line: 1, column: 18},
		},
		{
			name:    "single closing parenthesis with period",
			comment: "// ).",
			ok:      true,
		},
		{
			name:    "single closing parenthesis without period",
			comment: "// )",
			issue:   position{line: 1, column: 5},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, ok := checkText(tt.comment)
			if ok != tt.ok {
				t.Fatalf("Wrong result\n  expected: %v\n       got: %v", tt.ok, ok)
			}
			if pos.line != tt.issue.line {
				t.Fatalf("Wrong line\n  expected: %d\n       got: %d", tt.issue.line, pos.line)
			}
			if pos.column != tt.issue.column {
				t.Fatalf("Wrong column\n  expected: %d\n       got: %d", tt.issue.column, pos.column)
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

func TestRun(t *testing.T) {
	testFile := filepath.Join("testdata", "check", "main.go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse input file: %v", err)
	}

	testCases := []struct {
		name     string
		scope    Scope
		contains []string
	}{
		{
			name:     "scope: decl",
			scope:    DeclScope,
			contains: []string{"[DECL]"},
		},
		{
			name:     "scope: top",
			scope:    TopLevelScope,
			contains: []string{"[DECL]", "[TOP]"},
		},
		{
			name:     "scope: all",
			scope:    AllScope,
			contains: []string{"[DECL]", "[TOP]", "[ALL]"},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var expected int
			for _, c := range f.Comments {
				if strings.Contains(c.Text(), "[PASS]") {
					continue
				}
				for _, s := range tt.contains {
					if strings.Contains(c.Text(), s) {
						expected++
						break
					}
				}
			}
			issues, err := Run(f, fset, Settings{Scope: tt.scope})
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if len(issues) != expected {
				t.Fatalf("Wrong number of result issues\n  expected: %d\n       got: %d",
					expected, len(issues))
			}
		})
	}
}

func TestFix(t *testing.T) {
	testFile := filepath.Join("testdata", "check", "main.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file %s: %v", testFile, err)
	}
	content, err := ioutil.ReadFile(testFile) // nolint: gosec
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", testFile, err)
	}

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

	t.Run("scope: decl", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")

		fixed, err := Fix(testFile, file, fset, Settings{Scope: DeclScope})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: top", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")
		expected = strings.ReplaceAll(expected, "[TOP]", "[TOP].")

		fixed, err := Fix(testFile, file, fset, Settings{Scope: TopLevelScope})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: all", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")
		expected = strings.ReplaceAll(expected, "[TOP]", "[TOP].")
		expected = strings.ReplaceAll(expected, "[ALL]", "[ALL].")

		fixed, err := Fix(testFile, file, fset, Settings{Scope: AllScope})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})
}

func TestReplace(t *testing.T) {
	testFile := filepath.Join("testdata", "check", "main.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file %s: %v", testFile, err)
	}
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to check test file %s: %v", testFile, err)
	}
	mode := info.Mode()
	content, err := ioutil.ReadFile(testFile) // nolint: gosec
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", testFile, err)
	}

	t.Run("file not found", func(t *testing.T) {
		path := filepath.Join("testdata", "not-exists.go")
		err := Replace(path, nil, nil, Settings{})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("scope: decl", func(t *testing.T) {
		defer func() {
			ioutil.WriteFile(testFile, content, mode) // nolint: errcheck,gosec
		}()
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")

		if err := Replace(testFile, file, fset, Settings{Scope: DeclScope}); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := ioutil.ReadFile(testFile) // nolint: gosec
		if err != nil {
			t.Fatalf("Failed to read fixed file %s: %v", testFile, err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: top", func(t *testing.T) {
		defer func() {
			ioutil.WriteFile(testFile, content, mode) // nolint: errcheck,gosec
		}()
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")
		expected = strings.ReplaceAll(expected, "[TOP]", "[TOP].")
		expected = strings.ReplaceAll(expected, "[ALL]", "[ALL].")

		if err := Replace(testFile, file, fset, Settings{Scope: AllScope}); err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := ioutil.ReadFile(testFile) // nolint: gosec
		if err != nil {
			t.Fatalf("Failed to read fixed file %s: %v", testFile, err)
		}

		assertEqualContent(t, expected, string(fixed))
	})
}

func assertEqualContent(t *testing.T, expected, content string) {
	contentLines := strings.Split(content, "\n")
	expectedLines := strings.Split(expected, "\n")
	if len(contentLines) != len(expectedLines) {
		t.Fatalf("Invalid number of lines\n  expected: %d\n       got: %d",
			len(expectedLines), len(contentLines))
	}
	for i := range contentLines {
		// NOTE: This is a fix for Windows, not sure why this is happening
		result := strings.TrimRight(contentLines[i], "\r")
		exp := strings.TrimRight(expectedLines[i], "\r")
		if result != exp {
			t.Fatalf("Wrong line %d\n  expected: %s\n       got: %s",
				i, exp, result)
		}
	}
}
