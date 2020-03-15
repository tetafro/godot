package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"

	"github.com/tetafro/godot"
)

func main() {
	if len(os.Args) < 2 {
		fatal("Usage:\n  godot [FILES]")
	}
	input := os.Args[1:]

	var files []*ast.File
	fset := token.NewFileSet()

	for _, f := range input {
		stat, err := os.Stat(f)
		if os.IsNotExist(err) {
			fatal("File %s does not exist", f)
		}
		if stat.IsDir() {
			continue
		}
		if !strings.HasSuffix(f, ".go") {
			continue
		}

		file, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
		if err != nil {
			fatal("Failed to parse file %s: %v", f, err)
		}
		files = append(files, file)
	}

	for _, file := range files {
		msgs := godot.Run(file, fset)
		for _, msg := range msgs {
			fmt.Printf("%s: %s\n", msg.Message, msg.Pos)
		}
	}
}

func fatal(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
