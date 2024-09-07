package godot

import (
	"go/token"
	"testing"
)

func TestCheckPeriod(t *testing.T) {
	start := token.Position{
		Filename: "filename.go",
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	testCases := []struct {
		name    string
		comment comment
		issue   *Issue
	}{
		{
			name: "singleline text with period",
			comment: comment{
				lines: []string{"//Hello, world."},
				text:  "Hello, world.",
				start: start,
			},
			issue: nil,
		},
		{
			name: "singleline text with period and indentation",
			comment: comment{
				lines: []string{"//   Hello, world."},
				text:  "   Hello, world.",
				start: start,
			},
			issue: nil,
		},
		{
			name: "multiline text with period",
			comment: comment{
				lines: []string{"// Hello,", "// world."},
				text:  " Hello,\n world.",
				start: start,
			},
			issue: nil,
		},
		{
			name: "multiline text with period and empty lines",
			comment: comment{
				lines: []string{"/*", "Hello, world.", "*/"},
				text:  "\nHello, world.\n",
			},
			issue: nil,
		},
		{
			name: "singleline text with no period",
			comment: comment{
				lines: []string{"// Hello, world"},
				text:  " Hello, world",
				start: start,
			},
			issue: &Issue{
				Pos: token.Position{
					Filename: start.Filename,
					Line:     1,
					Column:   16,
				},
				Message:     noPeriodMessage,
				Replacement: "// Hello, world.",
			},
		},
		{
			name: "multiple text with no period",
			comment: comment{
				lines: []string{"/*", "Hello,", "world", "*/"},
				text:  "\nHello,\nworld\n",
				start: start,
			},
			issue: &Issue{
				Pos: token.Position{
					Filename: start.Filename,
					Line:     3,
					Column:   6,
				},
				Message:     noPeriodMessage,
				Replacement: "world.",
			},
		},
		{
			name: "question mark",
			comment: comment{
				lines: []string{"// Hello, world?"},
				text:  " Hello, world?",
				start: start,
			},
			issue: nil,
		},
		{
			name: "exclamation mark",
			comment: comment{
				lines: []string{"// Hello, world!"},
				text:  " Hello, world!",
				start: start,
			},
			issue: nil,
		},
		{
			name: "empty line",
			comment: comment{
				lines: []string{"//"},
				text:  "",
				start: start,
			},
			issue: nil,
		},
		{
			name: "empty lines",
			comment: comment{
				lines: []string{"/*", "", "", "*/"},
				text:  "\n\n",
				start: start,
			},
			issue: nil,
		},
		{
			name: "only spaces",
			comment: comment{
				lines: []string{"//   "},
				text:  "   ",
				start: start,
			},
			issue: nil,
		},
		{
			name: "mixed spaces",
			comment: comment{
				lines: []string{"//\t\t  "},
				text:  "\t\t  ",
				start: start,
			},
			issue: nil,
		},
		{
			name: "mixed spaces and newlines",
			comment: comment{
				lines: []string{"// \t\t \n\n\n  \n\t  "},
				text:  " \t\t \n\n\n  \n\t  ",
				start: start,
			},
			issue: nil,
		},
		{
			name: "cyrillic, with period",
			comment: comment{
				lines: []string{"// Кириллица."},
				text:  " Кириллица.",
				start: start,
			},
			issue: nil,
		},
		{
			name: "cyrillic, without period",
			comment: comment{
				lines: []string{"// Кириллица"},
				text:  " Кириллица",
				start: start,
			},
			issue: &Issue{
				Pos: token.Position{
					Filename: "filename.go",
					Offset:   0,
					Line:     1,
					Column:   22,
				},
				Message:     "Comment should end in a period",
				Replacement: "// Кириллица.",
			},
		},
		{
			name: "parenthesis, with period",
			comment: comment{
				lines: []string{"// Hello. (World.)"},
				text:  " Hello. (World.)",
				start: start,
			},
			issue: nil,
		},
		{
			name: "parenthesis, without period",
			comment: comment{
				lines: []string{"// Hello. (World)"},
				text:  " Hello. (World)",
				start: start,
			},
			issue: &Issue{
				Pos: token.Position{
					Filename: "filename.go",
					Offset:   0,
					Line:     1,
					Column:   18,
				},
				Message:     "Comment should end in a period",
				Replacement: "// Hello. (World).",
			},
		},
		{
			name: "single closing parenthesis with period",
			comment: comment{
				lines: []string{"//)."},
				text:  ").",
				start: start,
			},
			issue: nil,
		},
		{
			name: "single closing parenthesis without period",
			comment: comment{
				lines: []string{"//)"},
				text:  ")",
				start: start,
			},
			issue: &Issue{
				Pos: token.Position{
					Filename: "filename.go",
					Offset:   0,
					Line:     1,
					Column:   4,
				},
				Message:     "Comment should end in a period",
				Replacement: "//).",
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			issue := checkPeriod(tt.comment)
			switch {
			case tt.issue == nil && issue == nil:
				return
			case tt.issue == nil && issue != nil:
				t.Fatalf("Unexpected issue")
			case tt.issue != nil && issue == nil:
				t.Fatalf("Expected issue, got nil")
			case issue.Pos != tt.issue.Pos:
				t.Fatalf("Wrong position\n  expected: %+v\n       got: %+v",
					tt.issue.Pos, issue.Pos)
			case issue.Message != tt.issue.Message:
				t.Fatalf("Wrong message\n  expected: %s\n       got: %s",
					tt.issue.Message, issue.Message)
			case issue.Replacement != tt.issue.Replacement:
				t.Fatalf("Wrong replacement\n  expected: %s\n       got: %s",
					tt.issue.Replacement, issue.Replacement)
			}
		})
	}
}

func TestCheckCapital(t *testing.T) {
	start := token.Position{
		Filename: "filename.go",
		Offset:   0,
		Line:     1,
		Column:   1,
	}

	testCases := []struct {
		name    string
		comment comment
		issues  []Issue
	}{
		{
			name: "single sentence starting with a capital letter",
			comment: comment{
				lines: []string{"//Hello, world."},
				text:  "Hello, world.",
				start: start,
			},
		},
		{
			name: "single sentence starting with a lowercase letter",
			comment: comment{
				lines: []string{"// hello, world."},
				text:  " hello, world.",
				start: start,
			},
			issues: []Issue{
				{Pos: token.Position{Line: 1, Column: 4}},
			},
		},
		{
			name: "multiple sentences with mixed cases",
			comment: comment{
				lines: []string{
					"/* hello, world. Hello, world. hello? hello!",
					"",
					"hello, world. */",
				},
				text:  " hello, world. Hello, world. hello? hello!\n\nhello, world. ",
				start: start,
			},
			issues: []Issue{
				{Pos: token.Position{Line: 1, Column: 4}},
				{Pos: token.Position{Line: 1, Column: 32}},
				{Pos: token.Position{Line: 1, Column: 39}},
				{Pos: token.Position{Line: 3, Column: 1}},
			},
		},
		{
			name: "multiple sentences with mixed cases, declaration comment",
			comment: comment{
				lines: []string{
					"/* hello, world. Hello, world. hello? hello!",
					"",
					"hello, world. */",
				},
				text:  " hello, world. Hello, world. hello? hello!\n\nhello, world.",
				start: start,
				decl:  true,
			},
			issues: []Issue{
				{Pos: token.Position{Line: 1, Column: 32}},
				{Pos: token.Position{Line: 1, Column: 39}},
				{Pos: token.Position{Line: 3, Column: 1}},
			},
		},
		{
			name: "multiple sentences with cyrillic letters",
			comment: comment{
				lines: []string{"//Кириллица? кириллица!"},
				text:  "Кириллица? кириллица!",
				start: start,
			},
			issues: []Issue{
				{Pos: token.Position{Line: 1, Column: 23}},
			},
		},
		{
			name: "issue position column resolved from correct line",
			comment: comment{
				lines: []string{"// Кириллица.", "// Issue. here."},
				text:  " Кириллица.\n Issue. here.",
				start: start,
			},
			issues: []Issue{
				{Pos: token.Position{Line: 2, Column: 11}},
			},
		},
		{
			name: "sentence with leading spaces",
			comment: comment{
				lines: []string{"//    hello, world"},
				text:  "    hello, world",
				start: start,
			},
			issues: []Issue{
				{Pos: token.Position{Line: 1, Column: 7}},
			},
		},
		{
			name: "sentence with abbreviations",
			comment: comment{
				lines: []string{"//One two, i.e. hello, world, e.g. e. g. word and etc. word"},
				text:  "One two, i.e. hello, world, e.g. e. g. word and etc. word",
				start: start,
			},
			issues: nil,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			issues := checkCapital(tt.comment)
			if len(issues) != len(tt.issues) {
				t.Fatalf("Wrong number of issues\n  expected: %d\n       got: %d",
					len(tt.issues), len(issues))
			}
			for i := range issues {
				if issues[i].Pos.Line != tt.issues[i].Pos.Line {
					t.Fatalf("Wrong line\n  expected: %d\n       got: %d",
						tt.issues[i].Pos.Line, issues[i].Pos.Line)
				}
				if issues[i].Pos.Column != tt.issues[i].Pos.Column {
					t.Fatalf("Wrong column\n  expected: %d\n       got: %d",
						tt.issues[i].Pos.Column, issues[i].Pos.Column)
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
		{
			name:      "Test testing output",
			comment:   "// Output: true",
			isSpecial: true,
		},
		{
			name:      "Test multiline testing output",
			comment:   "// Output:\n// true\n// false",
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

func TestByteToRuneColumn(t *testing.T) {
	testCases := []struct {
		name  string
		str   string
		index int
		out   int
	}{
		{
			name:  "ascii symbols",
			str:   "hello, world",
			index: 5,
			out:   5,
		},
		{
			name:  "cyrillic symbols at the end",
			str:   "hello, мир",
			index: 5,
			out:   5,
		},
		{
			name:  "cyrillic symbols at the beginning",
			str:   "привет, world",
			index: 15,
			out:   9,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if out := byteToRuneColumn(tt.str, tt.index); out != tt.out {
				t.Fatalf("Wrong column\n  expected: %d\n       got: %d", tt.out, out)
			}
		})
	}
}

func TestRuneToByteColumn(t *testing.T) {
	testCases := []struct {
		name  string
		str   string
		index int
		out   int
	}{
		{
			name:  "ascii symbols",
			str:   "hello, world",
			index: 5,
			out:   5,
		},
		{
			name:  "cyrillic symbols at the end",
			str:   "hello, мир",
			index: 5,
			out:   5,
		},
		{
			name:  "cyrillic symbols at the beginning",
			str:   "привет, world",
			index: 9,
			out:   15,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if out := runeToByteColumn(tt.str, tt.index); out != tt.out {
				t.Fatalf("Wrong column\n  expected: %d\n       got: %d", tt.out, out)
			}
		})
	}
}
