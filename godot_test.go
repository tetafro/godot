package godot

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// testExclude is a test regexp to exclude lines that starts with @ symbol.
var testExclude = []string{"^ ?@"}

func TestRun(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		issues, err := Run(nil, nil, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(issues) > 0 {
			t.Fatal("Unexpected issues")
		}
	})

	t.Run("no comments", func(t *testing.T) {
		testFile := filepath.Join("testdata", "nocomments", "main.go")
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}

		issues, err := Run(f, fset, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(issues) > 0 {
			t.Fatal("Unexpected issues")
		}
	})

	t.Run("line directive", func(t *testing.T) {
		testFile := filepath.Join("testdata", "line", "main.go")
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}

		issues, err := Run(f, fset, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if len(issues) > 0 {
			t.Fatal("Unexpected issues")
		}
	})

	testFile := filepath.Join("testdata", "check", "main.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse input file: %v", err)
	}

	// Test invalid regexp
	_, err = Run(file, fset, Settings{
		Scope:   DeclScope,
		Exclude: []string{"["},
		Period:  true,
		Capital: true,
	})
	if err == nil {
		t.Fatalf("Expected error, got nil on regexp parsing")
	}

	testCases := []struct {
		name     string
		scope    Scope
		contains []string
	}{
		{
			name:  "scope: decl",
			scope: DeclScope,
			contains: []string{
				"[PERIOD_DECL]", "[CAPITAL_DECL]",
			},
		},
		{
			name:  "scope: top",
			scope: TopLevelScope,
			contains: []string{
				"[PERIOD_DECL]", "[CAPITAL_DECL]",
				"[PERIOD_TOP]", "[CAPITAL_TOP]",
			},
		},
		{
			name:  "scope: all",
			scope: AllScope,
			contains: []string{
				"[PERIOD_DECL]", "[CAPITAL_DECL]",
				"[PERIOD_TOP]", "[CAPITAL_TOP]",
				"[PERIOD_ALL]", "[CAPITAL_ALL]",
			},
		},
	}

	for _, tt := range testCases {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var expected int
			for _, c := range file.Comments {
				if strings.Contains(c.Text(), "[PASS]") {
					continue
				}
				for _, s := range tt.contains {
					if cnt := strings.Count(c.Text(), s); cnt > 0 {
						expected += cnt
					}
				}
			}
			issues, err := Run(file, fset, Settings{
				Scope:   tt.scope,
				Exclude: testExclude,
				Period:  true,
				Capital: true,
			})
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
	t.Run("file not found", func(t *testing.T) {
		testFile := filepath.Join("testdata", "not-exists.go")
		_, err := Fix(testFile, nil, nil, Settings{})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		testFile := filepath.Join("testdata", "empty", "main.go")

		fixed, err := Fix(testFile, nil, nil, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if fixed != nil {
			t.Fatalf("Unexpected result: %s", string(fixed))
		}
	})

	t.Run("no comments", func(t *testing.T) {
		testFile := filepath.Join("testdata", "nocomments", "main.go")
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read input file: %v", err)
		}

		fixed, err := Fix(testFile, f, fset, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertEqualContent(t, string(content), string(fixed))
	})

	t.Run("no code", func(t *testing.T) {
		testFile := filepath.Join("testdata", "nocode", "main.go")
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read input file: %v", err)
		}

		fixed, err := Fix(testFile, f, fset, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		assertEqualContent(t, string(content), string(fixed))
	})

	testFile := filepath.Join("testdata", "check", "main.go")
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse file %s: %v", testFile, err)
	}
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", testFile, err)
	}

	// Test invalid regexp
	_, err = Fix(testFile, file, fset, Settings{
		Scope:   DeclScope,
		Exclude: []string{"["},
		Period:  true,
		Capital: true,
	})
	if err == nil {
		t.Fatalf("Expected error, got nil on regexp parsing")
	}

	t.Run("scope: decl", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[PERIOD_DECL]", "[PERIOD_DECL].")
		expected = strings.ReplaceAll(expected, "non-capital-decl", "Non-capital-decl")

		fixed, err := Fix(testFile, file, fset, Settings{
			Scope:   DeclScope,
			Exclude: testExclude,
			Period:  true,
			Capital: true,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: top", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[PERIOD_DECL]", "[PERIOD_DECL].")
		expected = strings.ReplaceAll(expected, "[PERIOD_TOP]", "[PERIOD_TOP].")
		expected = strings.ReplaceAll(expected, "non-capital-decl", "Non-capital-decl")
		expected = strings.ReplaceAll(expected, "non-capital-top", "Non-capital-top")

		fixed, err := Fix(testFile, file, fset, Settings{
			Scope:   TopLevelScope,
			Exclude: testExclude,
			Period:  true,
			Capital: true,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: all", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[PERIOD_DECL]", "[PERIOD_DECL].")
		expected = strings.ReplaceAll(expected, "[PERIOD_TOP]", "[PERIOD_TOP].")
		expected = strings.ReplaceAll(expected, "[PERIOD_ALL]", "[PERIOD_ALL].")
		expected = strings.ReplaceAll(expected, "non-capital-decl", "Non-capital-decl")
		expected = strings.ReplaceAll(expected, "non-capital-top", "Non-capital-top")
		expected = strings.ReplaceAll(expected, "non-capital-all", "Non-capital-all")

		fixed, err := Fix(testFile, file, fset, Settings{
			Scope:   AllScope,
			Exclude: testExclude,
			Period:  true,
			Capital: true,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})
}

func TestReplace(t *testing.T) {
	t.Run("file not found", func(t *testing.T) {
		path := filepath.Join("testdata", "not-exists.go")
		err := Replace(path, nil, nil, Settings{})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		testFile := filepath.Join("testdata", "empty", "main.go")

		err := Replace(testFile, nil, nil, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
	})

	t.Run("no comments", func(t *testing.T) {
		testFile := filepath.Join("testdata", "nocomments", "main.go")
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read input file: %v", err)
		}

		err = Replace(testFile, f, fset, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read fixed file: %v", err)
		}
		assertEqualContent(t, string(content), string(fixed))
	})

	t.Run("no code", func(t *testing.T) {
		testFile := filepath.Join("testdata", "nocode", "main.go")
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
		if err != nil {
			t.Fatalf("Failed to parse input file: %v", err)
		}
		content, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read input file: %v", err)
		}

		err = Replace(testFile, f, fset, Settings{})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read fixed file: %v", err)
		}
		assertEqualContent(t, string(content), string(fixed))
	})

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
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read test file %s: %v", testFile, err)
	}

	// Test invalid regexp
	err = Replace(testFile, file, fset, Settings{
		Scope:   DeclScope,
		Exclude: []string{"["},
		Period:  true,
		Capital: true,
	})
	if err == nil {
		t.Fatalf("Expected error, got nil on regexp parsing")
	}

	t.Run("scope: decl", func(t *testing.T) {
		defer func() {
			os.WriteFile(testFile, content, mode)
		}()
		expected := strings.ReplaceAll(string(content), "[PERIOD_DECL]", "[PERIOD_DECL].")
		expected = strings.ReplaceAll(expected, "non-capital-decl", "Non-capital-decl")

		err := Replace(testFile, file, fset, Settings{
			Scope:   DeclScope,
			Exclude: testExclude,
			Period:  true,
			Capital: true,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read fixed file %s: %v", testFile, err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: top", func(t *testing.T) {
		defer func() {
			os.WriteFile(testFile, content, mode)
		}()
		expected := strings.ReplaceAll(string(content), "[PERIOD_DECL]", "[PERIOD_DECL].")
		expected = strings.ReplaceAll(expected, "[PERIOD_TOP]", "[PERIOD_TOP].")
		expected = strings.ReplaceAll(expected, "non-capital-decl", "Non-capital-decl")
		expected = strings.ReplaceAll(expected, "non-capital-top", "Non-capital-top")

		err := Replace(testFile, file, fset, Settings{
			Scope:   TopLevelScope,
			Exclude: testExclude,
			Period:  true,
			Capital: true,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := os.ReadFile(testFile)
		if err != nil {
			t.Fatalf("Failed to read fixed file %s: %v", testFile, err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: all", func(t *testing.T) {
		defer func() {
			os.WriteFile(testFile, content, mode)
		}()
		expected := strings.ReplaceAll(string(content), "[PERIOD_DECL]", "[PERIOD_DECL].")
		expected = strings.ReplaceAll(expected, "[PERIOD_TOP]", "[PERIOD_TOP].")
		expected = strings.ReplaceAll(expected, "[PERIOD_ALL]", "[PERIOD_ALL].")
		expected = strings.ReplaceAll(expected, "non-capital-decl", "Non-capital-decl")
		expected = strings.ReplaceAll(expected, "non-capital-top", "Non-capital-top")
		expected = strings.ReplaceAll(expected, "non-capital-all", "Non-capital-all")

		err := Replace(testFile, file, fset, Settings{
			Scope:   AllScope,
			Exclude: testExclude,
			Period:  true,
			Capital: true,
		})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		fixed, err := os.ReadFile(testFile)
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
			t.Fatalf("Wrong line %d\n  expected: '%s'\n       got: '%s'",
				i+1, exp, result)
		}
	}
}
