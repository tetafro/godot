// Package comment without a period [ALL]
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

//args: tagged comment without period [PASS]

// #tag hashtag comment without period [PASS]

/*
Multiline comment without a period [ALL]

*/

/*
Multiline comment with a period [PASS].
*/

/* One-line comment without a period [ALL] */

/* One-line comment with a period [PASS]. */

// Single-line comment without a period [ALL]

// Single-line comment with a period [PASS].

// Block comment [DEFAULT]
const (
	// Inside comment [DEFAULT]
	constant1 = "constant1"
	// Inside comment [PASS].
	constant2 = "constant2"
)

// Declaration comment without a period [DEFAULT]
type SimpleObject struct {
	// Exported field comment - always [PASS]
	Type string
	// Unexported field comment - always [PASS]
	secret int
}

// Declaration comment without a period, with an indented code example:
//   co := ComplexObject{}
//   fmt.Println(co) // [PASS]
type ComplexObject struct {
	// Exported field comment - always [PASS]
	Type string
	// Unexported field comment - always [PASS]
	secret int
}

// Declaration comment without a period, with a mixed indented code example:
// 	co := Message{}
// 	fmt.Println(co) // [PASS]
type Message struct {
	Type string
}

// Declaration multiline comment
// second line
// third line with a period [PASS].
func Sum(a, b int) int {
	// Inner comment - always [PASS]
	a++
	b++

	return a + b // inline comment - always [PASS]
}

// Declaration multiline comment
// second line
// third line without a period [DEFAULT]
func Mult(a, b int) int {
	return a * b
}

//export CgoExportedFunction [PASS]
func CgoExportedFunction(a, b int) int {
	return a + b
}

// Кириллица [DEFAULT]
func NonLatin() string {
	return "привет, мир"
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
	// Not a top level declaration [PASS]
	type thing struct {
		field string
	}
	t := thing{}
	println(t)
}

// Comment with a URL - http://example.com/[PASS]
