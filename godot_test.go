package godot

import (
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
			issues, err := Run(f, fset, Settings{Scope: tt.scope, Period: true})
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
		_, err := Fix(path, nil, nil, Settings{Period: true})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("empty file", func(t *testing.T) {
		path := filepath.Join("testdata", "empty.go")
		fixed, err := Fix(path, nil, nil, Settings{Period: true})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if fixed != nil {
			t.Fatalf("Unexpected result: %s", string(fixed))
		}
	})

	t.Run("scope: decl", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")

		fixed, err := Fix(testFile, file, fset, Settings{Scope: DeclScope, Period: true})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: top", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")
		expected = strings.ReplaceAll(expected, "[TOP]", "[TOP].")

		fixed, err := Fix(testFile, file, fset, Settings{Scope: TopLevelScope, Period: true})
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		assertEqualContent(t, expected, string(fixed))
	})

	t.Run("scope: all", func(t *testing.T) {
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")
		expected = strings.ReplaceAll(expected, "[TOP]", "[TOP].")
		expected = strings.ReplaceAll(expected, "[ALL]", "[ALL].")

		fixed, err := Fix(testFile, file, fset, Settings{Scope: AllScope, Period: true})
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
		err := Replace(path, nil, nil, Settings{Period: true})
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
	})

	t.Run("scope: decl", func(t *testing.T) {
		defer func() {
			ioutil.WriteFile(testFile, content, mode) // nolint: errcheck,gosec
		}()
		expected := strings.ReplaceAll(string(content), "[DECL]", "[DECL].")

		err := Replace(testFile, file, fset, Settings{Scope: DeclScope, Period: true})
		if err != nil {
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

		err := Replace(testFile, file, fset, Settings{Scope: AllScope, Period: true})
		if err != nil {
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
