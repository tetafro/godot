package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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

	for _, path := range input {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fatal("Path %s does not exist", path)
		}
		for f := range findFiles(path) {
			file, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
			if err != nil {
				fatal("Failed to parse file %s: %v", path, err)
			}
			files = append(files, file)
		}
	}

	for _, file := range files {
		msgs := godot.Run(file, fset)
		for _, msg := range msgs {
			fmt.Printf("%s: %s\n", msg.Message, msg.Pos)
		}
	}
}

func findFiles(root string) chan string {
	out := make(chan string)

	go func() {
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			sep := string(filepath.Separator)
			if strings.HasPrefix(path, "vendor"+sep) || strings.Contains(path, sep+"vendor"+sep) {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				out <- path
			}
			return nil
		})
		close(out)
	}()

	return out
}

func fatal(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
