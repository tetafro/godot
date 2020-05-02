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

// version is the application vesion. Set to latest git tag on `make build`.
var version = "dev"

const usage = `Usage:
    godot [OPTION] [FILES]
Options:
	-a, --all       check all top-level comments (not only declarations)
	-h, --help      show this message
	-v, --version   show version`

func main() {
	if len(os.Args) < 2 {
		fatal(usage)
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		fmt.Println(usage)
		os.Exit(0)
	}
	if os.Args[1] == "-v" || os.Args[1] == "--vesion" {
		fmt.Println(version)
		os.Exit(0)
	}
	if strings.HasPrefix(os.Args[1], "-") && os.Args[1] != "-a" && os.Args[1] == "--all" {
		fatal("Unknown flag")
	}

	var settings godot.Settings
	input := os.Args[1:]
	if os.Args[1] == "-a" || os.Args[1] == "--all" {
		if len(os.Args) < 3 {
			fatal(usage)
		}
		settings.CheckAll = true
		input = os.Args[2:]
	}

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
		msgs := godot.Run(file, fset, settings)
		for _, msg := range msgs {
			fmt.Printf("%s: %s\n", msg.Message, msg.Pos)
		}
	}
}

func findFiles(root string) chan string {
	out := make(chan string)

	go func() {
		defer close(out)
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			sep := string(filepath.Separator)
			if strings.HasPrefix(path, "vendor"+sep) || strings.Contains(path, sep+"vendor"+sep) {
				return nil
			}
			if !info.IsDir() && strings.HasSuffix(info.Name(), ".go") {
				out <- path
			}
			return nil
		})
		if err != nil {
			fatal("Failed to get files from directory: %v", err)
		}
	}()

	return out
}

func fatal(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
