// Package example is an example package [DECL].
package example

// Import comment [DECL].
import "fmt"

// Not gofmt-ed code [DECL].
const (
    one = 1

        two = 2
    
)

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
	// Const is an example constant. Top level declaration comment inside a block [DECL].
	Const = 1
)

// Top level comment for block of variables [DECL].
var (
	// Var is an example variable. Top level declaration comment inside a block [DECL].
	Var = 1
)

// Top level struct declaration comment [DECL].
type Thing struct {
	// Field is a struct field example; field declaration comment [ALL]
	Field
}

// Example is an example function. Top level function declaration comment [DECL].
func Example() { // top level inline comment [ALL]
	// Regular comment [ALL]

	// Declaration comment [ALL]
	prn := func() {
		// Nested comment [ALL]
		fmt.Println("hello, world")
	}

	prn() // Inline comment [ALL]
}
