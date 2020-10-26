package godot

import (
	"go/ast"
	"testing"
)

func TestCheckComments(t *testing.T) {
	// Check only case with empty input, other cases are checked in TestRun

	comments := []comment{
		{ast: nil},
		{ast: &ast.CommentGroup{List: nil}},
	}
	issues, err := checkComments(nil, comments, Settings{})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(issues) > 0 {
		t.Fatalf("Unexpected issues: %d", len(issues))
	}
}

func TestCheckPeriod(t *testing.T) {
	testCases := []struct {
		name  string
		text  string
		ok    bool
		issue position
	}{
		{
			name: "singleline text with period",
			text: "Hello, world.",
			ok:   true,
		},
		{
			name: "singleline text with period and indentation",
			text: " Hello, world.",
			ok:   true,
		},
		{
			name: "multiple text with period",
			text: "Hello,\nworld.",
			ok:   true,
		},
		{
			name: "multiple text with period and empty lines",
			text: "\nHello, world.\n",
			ok:   true,
		},
		{
			name:  "singleline text with no period",
			text:  "Hello, world",
			issue: position{line: 1, column: 13},
		},
		{
			name:  "multiple text with no period",
			text:  "\nHello,\nworld\n",
			issue: position{line: 3, column: 6},
		},
		{
			name: "question mark",
			text: "Hello, world?",
			ok:   true,
		},
		{
			name: "exclamation mark",
			text: "Hello, world!",
			ok:   true,
		},
		{
			name: "empty line",
			text: "",
			ok:   true,
		},
		{
			name: "empty lines",
			text: "\n\n",
			ok:   true,
		},
		{
			name: "only spaces",
			text: "   ",
			ok:   true,
		},
		{
			name: "mixed spaces",
			text: "\t\t  ",
			ok:   true,
		},
		{
			name: "mixed spaces and newlines",
			text: " \t\t \n\n\n  \n\t  ",
			ok:   true,
		},
		{
			name: "cyrillic, with period",
			text: "Кириллица.",
			ok:   true,
		},
		{
			name:  "cyrillic, without period",
			text:  "Кириллица",
			issue: position{line: 1, column: 10},
		},
		{
			name: "parenthesis, with period",
			text: "Hello. (World.)",
			ok:   true,
		},
		{
			name:  "parenthesis, without period",
			text:  "Hello. (World)",
			issue: position{line: 1, column: 15},
		},
		{
			name: "single closing parenthesis with period",
			text: ").",
			ok:   true,
		},
		{
			name:  "single closing parenthesis without period",
			text:  ")",
			issue: position{line: 1, column: 2},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			pos, ok := checkPeriod(tt.text)
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

func TestCheckCapital(t *testing.T) {
	testCases := []struct {
		name      string
		text      string
		skipFirst bool
		issues    []position
	}{
		{
			name:      "single sentence starting with a capital letter",
			text:      "Hello, world.",
			skipFirst: false,
		},
		{
			name:      "single sentence starting with a lowercase letter",
			text:      "hello, world.",
			skipFirst: false,
			issues: []position{
				{line: 1, column: 1},
			},
		},
		{
			name:      "multiple sentences with mixed cases",
			text:      "hello, world. Hello, world. hello? hello!\n\nhello, world.",
			skipFirst: false,
			issues: []position{
				{line: 1, column: 1},
				{line: 1, column: 29},
				{line: 1, column: 36},
				{line: 3, column: 1},
			},
		},
		{
			name:      "multiple sentences with mixed cases, first letter skipped",
			text:      "hello, world. Hello, world. hello? hello!\n\nhello, world.",
			skipFirst: true,
			issues: []position{
				{line: 1, column: 29},
				{line: 1, column: 36},
				{line: 3, column: 1},
			},
		},
		{
			name:      "multiple sentences with mixed cases",
			text:      "Кириллица? кириллица!",
			skipFirst: false,
			issues: []position{
				{line: 1, column: 12},
			},
		},
		{
			name:      "sentence with leading spaces",
			text:      "    hello, world",
			skipFirst: false,
			issues: []position{
				{line: 1, column: 5},
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCapital(tt.text, tt.skipFirst)
			if len(issues) != len(tt.issues) {
				t.Fatalf("Wrong number of issues\n  expected: %d\n       got: %d",
					len(tt.issues), len(issues))
			}
			for i := range issues {
				if issues[i].line != tt.issues[i].line {
					t.Fatalf("Wrong line\n  expected: %d\n       got: %d",
						tt.issues[i].line, issues[i].line)
				}
				if issues[i].column != tt.issues[i].column {
					t.Fatalf("Wrong column\n  expected: %d\n       got: %d",
						tt.issues[i].column, issues[i].column)
				}
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
