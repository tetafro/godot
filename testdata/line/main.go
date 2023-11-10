// This is a test for `//line` compiler directive. It can add references
// to other files, for parsed code. If it happens, it breaks consistency
// between the source code and the parsed files, which godot relies on.
package main

import "fmt"

func main() {
//line main.tpl:100
	fmt.Println("Template")
	// Bye!
}
