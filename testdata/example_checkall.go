// Package comment without a period FAIL
package example

//args: tagged comment without period PASS

// #tag hashtag comment without period PASS

/*
Multiline comment without a period FAIL

*/

/*
Multiline comment with a period PASS.
*/

/* One-line comment without a period FAIL */

/* One-line comment with a period PASS. */

// Single-line comment without a period FAIL

// Single-line comment with a period PASS.

// Declaration comment without a period FAIL
type ObjectX struct {
	// Exported field comment - always PASS
	Type string
	// Unexported field comment - always PASS
	secret int
}

// Declaration comment without a period, with an indented code example:
//   co := ComplexObject{}
//   fmt.Println(co) // PASS
type ComplexObjectX struct {
	// Exported field comment - always PASS
	Type string
	// Unexported field comment - always PASS
	secret int
}

// Declaration multiline comment
// second line
// third line with a period PASS.
func SumX(a, b int) int {
	// Inner comment - always PASS
	a++
	b++

	return a + b // inline comment - always PASS
}

// Declaration multiline comment
// second line
// third line without a period FAIL
func MultX(a, b int) int {
	return a * b
}

func noCommentX() {}

// Comment with a URL - http://example.com/PASS
