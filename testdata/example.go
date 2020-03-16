// Package example is an example pacakge.
package example

//args: tagged comment without period - ok

/*
Multiline comment without a period - bad

*/

/*
Multiline comment with a period - good.
*/

/* One-line comment without a period - bad */

/* One-line comment with a period - good. */

// Single-line comment without a period - bad

// Single-line comment with a period - good.

// Sum sums two integer elements
// second line
// third line with period - good.
func Sum(a, b int) int {
	return a + b // result
}

// Sum multiplies two integer elements
// second line
// third line without period - bad
func Mult(a, b int) int {
	return a * b // result
}

func nocomment() {}

// Comment with a URL - ok http://example.com
