// Package example is an example package [DECL].
package example

// Import comment [DECL].
import "fmt"

// Top level one-line comment [TOP].

// Top level
// multiline
// comment [TOP].

/* Top level one-line comment block [TOP]. */

/*
Top level
multiline
comment block [TOP].
*/

// Top level comment for block of constants [DECL].
const (
	// Const1 is an example constant. Top level declaration comment inside a block [DECL].
	Const1 = 1
)

// Top level comment for block of variables [DECL].
var (
	// Var1 is an example variable. Top level declaration comment inside a block [DECL].
	Var1 = 1
)

// Top level struct declaration comment [DECL].
type Thing struct {
	// Field is a struct field example; field declaration comment [NONE]
	Field
}

// Example is an example function. Top level function declaration comment [DECL].
func Example() { // top level inline comment [NONE]
	// Regular comment [NONE]

	// Declaration comment [NONE]
	prn := func() {
		// Nested comment [NONE]
		fmt.Println("hello, world")
	}

	prn() // inline comment [NONE]
}
