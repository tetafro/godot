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

// version is the application version. It is set to the latest git tag in CI.
var version = "master"

// nolint: lll
const usage = `Usage:
	godot [OPTION] [FILES]
Options:
	-s, --scope     set scope for check (decl for top level declaration comments, top for top level comments, all for all comments)
	-f, --fix       fix issues, and print fixed version to stdout
	-h, --help      show this message
	-v, --version   show version
	-w, --write     fix issues, and write result to original file`

type arguments struct {
	help    bool
	version bool
	fix     bool
	write   bool
	scope   string
	files   []string
}

func main() {
	args, err := readArgs()
	if err != nil {
		fatalf("Error: %v", err)
	}

	if args.help {
		fmt.Println(usage)
		os.Exit(0)
	}
	if args.version {
		fmt.Println(version)
		os.Exit(0)
	}
	if len(args.files) == 0 {
		fatalf(usage)
	}

	settings := godot.Settings{
		Scope: godot.Scope(args.scope),
	}

	var paths []string
	var files []*ast.File
	fset := token.NewFileSet()

	for _, path := range args.files {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fatalf("Path '%s' does not exist", path)
		}
		for f := range findFiles(path) {
			file, err := parser.ParseFile(fset, f, nil, parser.ParseComments)
			if err != nil {
				fatalf("Failed to parse file '%s': %v", path, err)
			}
			files = append(files, file)
			paths = append(paths, f)
		}
	}

	for i := range files {
		switch {
		case args.fix:
			fixed, err := godot.Fix(paths[i], files[i], fset, settings)
			if err != nil {
				fatalf("Failed to autofix file '%s': %v", paths[i], err)
			}
			fmt.Print(string(fixed))
		case args.write:
			if err := godot.Replace(paths[i], files[i], fset, settings); err != nil {
				fatalf("Failed to rewrite file '%s': %v", paths[i], err)
			}
		default:
			issues := godot.Run(files[i], fset, settings)
			for _, iss := range issues {
				fmt.Printf("%s: %s\n", iss.Message, iss.Pos)
			}
		}
	}
}

func readArgs() (args arguments, err error) {
	if len(os.Args) < 2 { // nolint: gomnd
		return arguments{}, fmt.Errorf("not enough arguments")
	}

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		if !strings.HasPrefix(arg, "-") {
			args.files = append(args.files, arg)
			continue
		}

		switch arg {
		case "-h", "--help":
			args.help = true
		case "-v", "--version":
			args.version = true
		case "-s", "--scope":
			// Next argument must be scope value
			if len(os.Args) < i+2 {
				return arguments{}, fmt.Errorf("empty scope")
			}
			arg = os.Args[i+1]
			i++

			switch arg {
			case string(godot.DeclScope),
				string(godot.TopLevelScope):
				args.scope = arg
			default:
				return arguments{}, fmt.Errorf("unknown scope '%s'", arg)
			}
		case "-f", "--fix":
			args.fix = true
		case "-w", "--write":
			args.write = true
		default:
			return arguments{}, fmt.Errorf("unknown flag '%s'", arg)
		}
	}
	return args, nil
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
			fatalf("Failed to get files from directory: %v", err)
		}
	}()

	return out
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
	os.Exit(1)
}
