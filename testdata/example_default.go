// Package comment without a period PASS
package example

//args: tagged comment without period PASS

// #tag hashtag comment without period PASS

/*
Multiline comment without a period PASS

*/

/*
Multiline comment with a period PASS.
*/

/* One-line comment without a period PASS */

/* One-line comment with a period PASS. */

// Single-line comment without a period PASS

// Single-line comment with a period PASS.

// Declaration comment without a period FAIL
type SimpleObject struct {
	// Exported field comment - always PASS
	Type string
	// Unexported field comment - always PASS
	secret int
}

// Declaration comment without a period, with an indented code example:
//   co := ComplexObject{}
//   fmt.Println(co) // PASS
type ComplexObject struct {
	// Exported field comment - always PASS
	Type string
	// Unexported field comment - always PASS
	secret int
}

// Declaration multiline comment
// second line
// third line with a period PASS.
func Sum(a, b int) int {
	// Inner comment - always PASS
	a++
	b++

	return a + b // inline comment - always PASS
}

// Declaration multiline comment
// second line
// third line without a period FAIL
func Mult(a, b int) int {
	return a * b
}

func noComment() {}

// Comment with a URL - http://example.com/PASS
