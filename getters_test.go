package godot

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
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

	pf, err := newParsedFile(file, fset)
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
			comments := pf.getComments(tt.scope, nil)
			var expected int
			for _, c := range comments {
				if linesContain(c.lines, "[NONE]") {
					continue
				}
				for _, s := range tt.contains {
					if strings.Contains(c.text, s) {
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

	t.Run("try to get comments from cgo generated file", func(t *testing.T) {
		testFile := filepath.Join("testdata", "get", "cgo.go")
		fset := token.NewFileSet()
		file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}

		pf, err := newParsedFile(file, fset)
		if pf != nil {
			t.Fatalf("Unexpected file content")
		}
		if !errors.Is(err, errUnsuitableInput) {
			t.Fatalf(
				"Unexpected error:\n  expected: %v\n       got: %v",
				errUnsuitableInput, err,
			)
		}
	})
}

func TestGetText(t *testing.T) {
	testCases := []struct {
		name    string
		comment *ast.CommentGroup
		text    string
		exclude *regexp.Regexp
	}{
		{
			name: "regular text",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// Hello, world"},
			}},
			text: " Hello, world",
		},
		{
			name: "regular text without indentation",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "//Hello, world"},
			}},
			text: "Hello, world",
		},
		{
			name: "empty singleline comment",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "//"},
			}},
			text: "",
		},
		{
			name: "empty multiline comment",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/**/"},
			}},
			text: "",
		},
		{
			name: "regular text in multiline block",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/*\nHello, world\n*/"},
			}},
			text: "\nHello, world\n",
		},
		{
			name: "block of singleline comments with regular text",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// One"},
				{Text: "// Two"},
				{Text: "// Three"},
			}},
			text: " One\n Two\n Three",
		},
		{
			name: "block of singleline comments with empty and special lines",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// One"},
				{Text: "//"},
				{Text: "//  fmt.Println(s)"},
				{Text: "// Two"},
				{Text: "// #nosec"},
				{Text: "// Three"},
				{Text: "// +k8s:deepcopy-gen=package"},
				{Text: "// +nolint: gosec"},
			}},
			text: " One\n" +
				"\n" +
				"<godotSpecialReplacer>\n" +
				" Two\n" +
				"<godotSpecialReplacer>\n" +
				" Three\n" +
				"<godotSpecialReplacer>\n" +
				"<godotSpecialReplacer>",
		},
		{
			name: "block of singleline comments with a url at the end",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// Read more"},
				{Text: "// http://example.com"},
			}},
			text: " Read more\n<godotSpecialReplacer>",
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
			text: "\n" +
				"Example:\n" +
				"<godotSpecialReplacer>\n" +
				"<godotSpecialReplacer>\n" +
				"",
		},
		{
			name:    "empty comment group",
			comment: &ast.CommentGroup{List: []*ast.Comment{}},
			text:    "",
		},
		{
			name: "single excluded line",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "// Hello, world."},
			}},
			text:    "<godotSpecialReplacer>",
			exclude: regexp.MustCompile("Hello"),
		},
		{
			name: "excluded line in the middle",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/*\n" +
					"Read more:\n" +
					"@intenal.link\n" +
					"Thanks." +
					"*/"},
			}},
			text: "\n" +
				"Read more:\n" +
				"<godotSpecialReplacer>\n" +
				"Thanks." +
				"",
			exclude: regexp.MustCompile("^@.+"),
		},
		{
			name: "excluded line at the end",
			comment: &ast.CommentGroup{List: []*ast.Comment{
				{Text: "/* Read more:\n" +
					"@intenal.link */"},
			}},
			text: " Read more:\n" +
				"<godotSpecialReplacer>",
			exclude: regexp.MustCompile("^@.+"),
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var re []*regexp.Regexp
			if tt.exclude != nil {
				re = []*regexp.Regexp{tt.exclude}
			}
			if text := getText(tt.comment, re); text != tt.text {
				t.Fatalf("Wrong text\n  expected: '%s'\n       got: '%s'", tt.text, text)
			}
		})
	}
}

func linesContain(lines []string, s string) bool {
	for _, ln := range lines {
		if strings.Contains(ln, s) {
			return true
		}
	}
	return false
}
