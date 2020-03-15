package godot

import "testing"

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
