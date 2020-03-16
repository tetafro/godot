package godot

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"testing"
)

var testFile = filepath.Join("testdata", "example.go")

func TestCheckComment(t *testing.T) {
	testCases := []struct {
		name    string
		comment string
		ok      bool
		line    int
	}{
		// Single line comments
		{
			name:    "singleline comment: ok",
			comment: "// Hello, world.",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: no period",
			comment: "// Hello, world",
			ok:      false,
			line:    0,
		},
		{
			name:    "singleline comment: question mark",
			comment: "// Hello, world?",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: exclamation mark",
			comment: "// Hello, world!",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: code example without period",
			comment: "//  x == y",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: code example without period and long indentation",
			comment: "//       x == y",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: code example without period and tab indentation",
			comment: "//\tx == y",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: empty line",
			comment: "//",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: empty line with spaces",
			comment: "//      ",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: without indentation and with period",
			comment: "//hello, world.",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: without indentation and without period",
			comment: "//hello, world",
			ok:      false,
			line:    0,
		},
		{
			name:    "singleline comment: nolint mark without period",
			comment: "// nolint: test",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: nolint mark without indentation without period",
			comment: "//nolint: test",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: build tags",
			comment: "// +build !linux",
			ok:      true,
			line:    0,
		},
		{
			name:    "singleline comment: url at the end of line",
			comment: "// Read more: http://example.com/",
			ok:      true,
			line:    0,
		},
		// Multiline comments
		{
			name:    "multiline comment: ok",
			comment: "/*\n" + "Hello, world.\n" + "*/",
			ok:      true,
			line:    1,
		},
		{
			name:    "multiline comment: no period",
			comment: "/*\n" + "Hello, world\n" + "*/",
			ok:      false,
			line:    1,
		},
		{
			name:    "multiline comment: question mark",
			comment: "/*\n" + "Hello, world?\n" + "*/",
			ok:      true,
			line:    1,
		},
		{
			name:    "multiline comment: exclamation mark",
			comment: "/*\n" + "Hello, world!\n" + "*/",
			ok:      true,
			line:    1,
		},
		{
			name:    "multiline comment: code example without period",
			comment: "/*\n" + "  x == y\n" + "*/",
			ok:      true,
			line:    1,
		},
		{
			name:    "multiline comment: code example without period and long indentation",
			comment: "/*\n" + "       x == y\n" + "*/",
			ok:      true,
			line:    1,
		},
		{
			name:    "multiline comment: code example without period and tab indentation",
			comment: "/*\n" + "\tx == y\n" + "*/",
			ok:      true,
			line:    1,
		},
		{
			name:    "multiline comment: empty line",
			comment: "/**/",
			ok:      true,
			line:    0,
		},
		{
			name:    "multiline comment: empty line inside",
			comment: "/*\n" + "\n" + "*/",
			ok:      true,
			line:    0,
		},
		{
			name:    "multiline comment: empty line with spaces",
			comment: "/*\n" + "    \n" + "*/",
			ok:      true,
			line:    0,
		},
		{
			name:    "multiline comment: closing with whitespaces",
			comment: "/*\n" + "    \n" + "   */",
			ok:      true,
			line:    0,
		},
		{
			name:    "multiline comment: one-liner with period",
			comment: "/* Hello, world. */",
			ok:      true,
			line:    0,
		},
		{
			name:    "multiline comment: one-liner with period and without indentation",
			comment: "/*Hello, world.*/",
			ok:      true,
			line:    0,
		},
		{
			name:    "multiline comment: one-liner without period and without indentation",
			comment: "/*Hello, world*/",
			ok:      false,
			line:    0,
		},
		{
			name:    "multiline comment: long comment",
			comment: "/*\n" + "\n" + "   \n" + "Hello, world\n" + "\n" + "  \n" + "*/",
			ok:      false,
			line:    3,
		},
		{
			name:    "multiline comment: url at the end of line",
			comment: "/*\n" + "Read more: http://example.com/" + "*/",
			ok:      true,
			line:    1,
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			line, ok := checkComment(tt.comment)
			if ok != tt.ok {
				t.Fatalf("Wrong result, expected %v, got %v", tt.ok, ok)
			}
			if line != tt.line {
				t.Fatalf("Wrong line, expected %d, got %d", tt.line, line)
			}
		})
	}
}

func TestIntegration(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file %s: %v", testFile, err)
	}

	t.Run("check all", func(t *testing.T) {
		expected := []Message{
			{
				Message: "Top level comment should end in a period",
				Pos: token.Position{
					Filename: testFile,
					Column:   1,
					Line:     7,
				},
			},
			{
				Message: "Top level comment should end in a period",
				Pos: token.Position{
					Filename: testFile,
					Column:   1,
					Line:     15,
				},
			},
			{
				Message: "Top level comment should end in a period",
				Pos: token.Position{
					Filename: testFile,
					Column:   1,
					Line:     19,
				},
			},
			{
				Message: "Top level comment should end in a period",
				Pos: token.Position{
					Filename: testFile,
					Column:   1,
					Line:     32,
				},
			},
		}

		msgs := Run(file, fset, Settings{CheckAll: true})
		if len(msgs) != len(expected) {
			t.Fatalf("Invalid number of result messages\n  expected: %d\n       got: %d",
				len(expected), len(msgs))
		}
		for i := range msgs {
			if msgs[i].Pos.Filename != expected[i].Pos.Filename ||
				msgs[i].Pos.Line != expected[i].Pos.Line {
				t.Fatalf("Unexpected position\n  expected %s\n       got %s",
					expected[i].Pos, msgs[i].Pos)
			}
		}
	})

	t.Run("check declarations", func(t *testing.T) {
		expected := []Message{
			{
				Message: "Top level comment should end in a period",
				Pos: token.Position{
					Filename: testFile,
					Column:   1,
					Line:     32,
				},
			},
		}

		msgs := Run(file, fset, Settings{CheckAll: false})
		if len(msgs) != len(expected) {
			t.Fatalf("Invalid number of result messages\n  expected: %d\n       got: %d",
				len(expected), len(msgs))
		}
		for i := range msgs {
			if msgs[i].Pos.Filename != expected[i].Pos.Filename ||
				msgs[i].Pos.Line != expected[i].Pos.Line {
				t.Fatalf("Unexpected position\n  expected %s\n       got %s",
					expected[i].Pos, msgs[i].Pos)
			}
		}
	})
}
