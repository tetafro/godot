// Package comment without a period [PERIOD_TOP]
package example

/*
#include <stdio.h>
#include <stdlib.h>

void myprint(char* s) {
	printf("%d\n", s);
}

# [PASS]
*/
import (
	"C"
	"unsafe"
)

// Not gofmt-ed code [PASS].
const (
    one = 1

        two = 2
    
)

//args: tagged comment without period [PASS]

// #tag hashtag comment without period [PASS]

/*
non-capital-top [CAPITAL_TOP].
Multiline comment without a period [PERIOD_TOP]

*/

/*
Multiline comment with a period [PASS].
*/

/* One-line comment without a period [PERIOD_TOP] */

/* One-line comment with a period [PASS]. */

// Single-line comment without a period [PERIOD_TOP]

// Single-line comment with a period [PASS].

// Mixed block of comments,
/*
period must be here [PERIOD_TOP]
*/

/* Mixed block of comments,
*/
// period must be here [PERIOD_TOP]

/*
// Comment inside comment [PERIOD_TOP]
*/

// Block comment [PERIOD_DECL]
const (
	// Inside comment [PERIOD_DECL]
	constant1 = "constant1"
	// Inside comment [PASS].
	constant2 = "constant2"
)

// Declaration comment without a period [PERIOD_DECL]
type SimpleObject struct {
	// Exported field comment [PERIOD_ALL]
	Type string
	// Unexported field comment [PERIOD_ALL]
	secret int
}

// Declaration comment without a period, with an indented code example:
//   co := ComplexObject{}
//   fmt.Println(co) // [PASS]
type ComplexObject struct {
	// Exported field comment [PERIOD_ALL]
	Type string
	// Unexported field comment [PERIOD_ALL]
	secret int
}

// Declaration comment without a period, with a mixed indented code example:
// 	co := Message{}
// 	fmt.Println(co) // [PASS]
type Message struct {
	Type string
}

// Generic type [PASS].
type Array[T int64 | float64] struct {
	Elements []T
}

// Declaration multiline comment
// second line
// third line with a period [PASS].
func Sum(a, b int) int {
	// Inner comment [PERIOD_ALL]
	a++
	b++

	return a + b // Inline comment [PERIOD_ALL]
}

// Declaration multiline comment
// second line
// third line without a period [PERIOD_DECL]
func Mult(a, b int) int {
	return a * b
}

//export CgoExportedFunction [PASS]
func CgoExportedFunction(a, b int) int {
	return a + b
}

// Кириллица [PERIOD_DECL]
func NonLatin() string {
	// Тест: Mixed ASCII and non-ASCII chars.
	return "привет, мир"
}

// Asian period [PASS]。
func Asian() {
	return "日本語"
}

// Comment. (Parenthesis [PASS].)
func Parenthesis() string {
	return "привет, мир"
}

func noComment() {
	cs := C.CString("Hello from stdio\n")
	C.myprint(cs)
	C.free(unsafe.Pointer(cs))
}

func inside() {
	// Not a top level declaration [PERIOD_ALL]
	type thing struct {
		field string
	}
	t := thing{} // Inline comment [PERIOD_ALL]
	println(t)
	// @Comment without a period excluded by regexp pattern [PASS]
}

// nonCapital is a function. non-capital-decl first letter [CAPITAL_DECL].
func nonCapital() int {
	// Test abbreviation (e.g. like this) [PASS].
	x := 10

	// non-capital-all [CAPITAL_ALL].
	return x // non-capital-all [CAPITAL_ALL].
}

// Comment with a URL - http://example.com/[PASS]

// Multiline comment with a URL
// http://example.com/[PASS]

// @Comment without a period excluded by regexp pattern [PASS]
