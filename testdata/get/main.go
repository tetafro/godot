// Package example is an example package [DEFAULT].
package example

// Import comment [DEFAULT].
import "fmt"

// Top level one-line comment [ALL].

// Top level
// multiline
// comment [ALL].

/* Top level one-line comment block [ALL]. */

/*
Top level
multiline
comment block [ALL].
*/

// Top level comment for block of constants [DEFAULT].
const (
	// Const1 is an example constant. Top level declaration comment inside a block [DEFAULT].
	Const1 = 1
)

// Top level comment for block of variables [DEFAULT].
var (
	// Var1 is an example variable. Top level declaration comment inside a block [DEFAULT].
	Var1 = 1
)

// Top level struct declaration comment [DEFAULT].
type Thing struct {
	// Field is a struct field example; field declaration comment [NONE]
	Field
}

// Example is an example function. Top level function declaration comment [DEFAULT].
func Example() { // top level inline comment [NONE]
	// Regular comment [NONE]

	// Declaration comment [NONE]
	prn := func() {
		// Nested comment [NONE]
		fmt.Println("hello, world")
	}

	prn() // inline comment [NONE]
}
