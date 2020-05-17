package godot

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckComment(t *testing.T) {
	testCases := []struct {
		name    string
		comment string
		ok      bool
		pos     position
	}{
		// Single line comments
		{
			name:    "singleline comment: ok",
			comment: "// Hello, world.",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: no period",
			comment: "// Hello, world",
			ok:      false,
			pos:     position{line: 0, column: 15},
		},
		{
			name:    "singleline comment: question mark",
			comment: "// Hello, world?",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: exclamation mark",
			comment: "// Hello, world!",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: code example without period",
			comment: "//  x == y",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: code example without period and long indentation",
			comment: "//       x == y",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: code example without period and tab indentation",
			comment: "//\tx == y",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: code example without period and mixed indentation",
			comment: "// \tx == y",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: empty line",
			comment: "//",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: empty line with spaces",
			comment: "//      ",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: without indentation and with period",
			comment: "//hello, world.",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: without indentation and without period",
			comment: "//hello, world",
			ok:      false,
			pos:     position{line: 0, column: 14},
		},
		{
			name:    "singleline comment: nolint mark without period",
			comment: "// nolint: test",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: nolint mark without indentation without period",
			comment: "//nolint: test",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: build tags",
			comment: "// +build !linux",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: cgo exported function",
			comment: "//export FuncName",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: url at the end of line",
			comment: "// Read more: http://example.com/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: cyrillic, with period",
			comment: "// Кириллица.",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "singleline comment: cyrillic, without period",
			comment: "// Кириллица",
			ok:      false,
			pos:     position{line: 0, column: 12},
		},
		// Multiline comments
		{
			name:    "multiline comment: ok",
			comment: "/*\n" + "Hello, world.\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: no period",
			comment: "/*\n" + "Hello, world\n" + "*/",
			ok:      false,
			pos:     position{line: 1, column: 12},
		},
		{
			name:    "multiline comment: question mark",
			comment: "/*\n" + "Hello, world?\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: exclamation mark",
			comment: "/*\n" + "Hello, world!\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: code example without period",
			comment: "/*\n" + "  x == y\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: code example without period and long indentation",
			comment: "/*\n" + "       x == y\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: code example without period and tab indentation",
			comment: "/*\n" + "\tx == y\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: empty line",
			comment: "/**/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: empty line inside",
			comment: "/*\n" + "\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: empty line with spaces",
			comment: "/*\n" + "    \n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: closing with whitespaces",
			comment: "/*\n" + "    \n" + "   */",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: one-liner with period",
			comment: "/* Hello, world. */",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: one-liner with period and without indentation",
			comment: "/*Hello, world.*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: one-liner without period and without indentation",
			comment: "/*Hello, world*/",
			ok:      false,
			pos:     position{line: 0, column: 14},
		},
		{
			name:    "multiline comment: long comment",
			comment: "/*\n" + "\n" + "   \n" + "Hello, world\n" + "\n" + "  \n" + "*/",
			ok:      false,
			pos:     position{line: 3, column: 12},
		},
		{
			name:    "multiline comment: url at the end of line",
			comment: "/*\n" + "Read more: http://example.com/\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: cyrillic, with period",
			comment: "/*\n" + "Кириллица.\n" + "*/",
			ok:      true,
			pos:     position{},
		},
		{
			name:    "multiline comment: cyrillic, without period",
			comment: "/*\n" + "Кириллица\n" + "*/",
			ok:      false,
			pos:     position{line: 1, column: 9},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, ok := checkComment(tt.comment)
			if ok != tt.ok {
				t.Fatalf("Wrong result, expected %v, got %v", tt.ok, ok)
			}
			if pos.line != tt.pos.line {
				t.Fatalf("Wrong line, expected %d, got %d", tt.pos.line, pos.line)
			}
			if pos.column != tt.pos.column {
				t.Fatalf("Wrong column, expected %d, got %d", tt.pos.column, pos.column)
			}
		})
	}
}

func TestMakeReplacement(t *testing.T) {
	testCases := []struct {
		name        string
		comment     string
		pos         position
		replacement string
	}{
		{
			name:        "singleline comment",
			comment:     "// Hello, world",
			pos:         position{line: 0, column: 15},
			replacement: "// Hello, world.",
		},
		{
			name:        "short singleline comment",
			comment:     "//x",
			pos:         position{line: 0, column: 3},
			replacement: "//x.",
		},
		{
			name:        "cyrillic singleline comment",
			comment:     "// Привет, мир",
			pos:         position{line: 0, column: 14},
			replacement: "// Привет, мир.",
		},
		{
			name:        "multiline comment",
			comment:     "/*\n" + "Hello, world\n" + "*/",
			pos:         position{line: 1, column: 12},
			replacement: "Hello, world.",
		},
		{
			name:        "short multiline comment",
			comment:     "/*\n" + "x\n" + "*/",
			pos:         position{line: 1, column: 1},
			replacement: "x.",
		},
		{
			name:        "multiline comment in one line",
			comment:     "/* Hello, world */",
			pos:         position{line: 0, column: 15},
			replacement: "/* Hello, world. */",
		},
		{
			name:        "cyrillic multiline comment",
			comment:     "/* Привет, мир */",
			pos:         position{line: 0, column: 14},
			replacement: "/* Привет, мир. */",
		},
		{
			name:        "invalid line",
			comment:     "// Привет, мир",
			pos:         position{line: 100, column: 0},
			replacement: "// Привет, мир",
		},
		{
			name:        "invalid column",
			comment:     "// Привет, мир",
			pos:         position{line: 0, column: 100},
			replacement: "// Привет, мир",
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			replacement := makeReplacement(tt.comment, tt.pos)
			if replacement != tt.replacement {
				t.Fatalf(
					"Wrong replacement\n  expected: %s\n       got: %s",
					tt.replacement, replacement,
				)
			}
		})
	}
}

func TestRunIntegration(t *testing.T) {
	t.Run("default check", func(t *testing.T) {
		var testFile = filepath.Join("testdata", "default", "check", "main.go")
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
		var testFile = filepath.Join("testdata", "checkall", "check", "main.go")
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

func TestFixIntegration(t *testing.T) {
	t.Run("default check", func(t *testing.T) {
		var testFile = filepath.Join("testdata", "default", "check", "main.go")
		var expectedFile = filepath.Join("testdata", "default", "result", "main.go")
		expected, err := ioutil.ReadFile(expectedFile) // nolint: gosec
		if err != nil {
			t.Fatalf("Failed to read test file %s: %v", expected, err)
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", testFile, err)
		}

		fixed, err := Fix(testFile, file, fset, Settings{CheckAll: false})
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

	t.Run("check all", func(t *testing.T) {
		var testFile = filepath.Join("testdata", "checkall", "check", "main.go")
		var expectedFile = filepath.Join("testdata", "checkall", "result", "main.go")
		expected, err := ioutil.ReadFile(expectedFile) // nolint: gosec
		if err != nil {
			t.Fatalf("Failed to read test file %s: %v", expected, err)
		}

		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse file %s: %v", testFile, err)
		}

		fixed, err := Fix(testFile, file, fset, Settings{CheckAll: true})
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
			// TODO: This is a fix for Windows, not sure why this is happening
			result := strings.TrimRight(fixedLines[i], "\r")
			exp := strings.TrimRight(expectedLines[i], "\r")
			if result != exp {
				t.Fatalf("Wrong line %d in fixed file\n  expected: %s\n       got: %s",
					i, exp, result)
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
