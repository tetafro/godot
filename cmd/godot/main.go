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
    -c, --capital   check that sentences start with a capital letter
    -s, --scope     set scope for check
                    decl - for top level declaration comments
                    top  - for top level comments (default)
                    all  - for all comments
    -f, --fix       fix issues, and print fixed version to stdout
    -h, --help      show this message
    -v, --version   show version
    -w, --write     fix issues, and write result to original file`

// nolint:maligned
type arguments struct {
	help    bool
	version bool
	fix     bool
	write   bool
	scope   string
	files   []string
	capital bool
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

	settings := godot.Settings{
		Scope:   godot.Scope(args.scope),
		Period:  true,
		Capital: args.capital,
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
			issues, err := godot.Run(files[i], fset, settings)
			if err != nil {
				fatalf("Failed to run linter on file '%s': %v", paths[i], err)
			}
			for _, iss := range issues {
				fmt.Printf("%s: %s\n", iss.Message, iss.Pos)
			}
		}
	}
}

// nolint: funlen
func readArgs() (args arguments, err error) {
	if len(os.Args) < 2 { // nolint: gomnd
		return arguments{}, fmt.Errorf("not enough arguments")
	}

	// Split `--arg=x` arguments
	input := make([]string, 0, len(os.Args)-1)
	for i := 1; i < len(os.Args); i++ {
		splitted := strings.Split(os.Args[i], "=")
		if len(splitted) > 2 { // nolint: gomnd
			return arguments{}, fmt.Errorf("invalid argument '%s'", os.Args[i])
		}
		input = append(input, splitted...)
	}

	for i := 0; i < len(input); i++ {
		arg := input[i]
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
			if len(input) < i+2 {
				return arguments{}, fmt.Errorf("empty scope")
			}
			arg = input[i+1]
			i++

			switch arg {
			case string(godot.DeclScope),
				string(godot.TopLevelScope),
				string(godot.AllScope):
				args.scope = arg
			default:
				return arguments{}, fmt.Errorf("unknown scope '%s'", arg)
			}
		case "-f", "--fix":
			args.fix = true
		case "-w", "--write":
			args.write = true
		case "-c", "--capital":
			args.capital = true
		default:
			return arguments{}, fmt.Errorf("unknown flag '%s'", arg)
		}
	}

	if args.scope == "" {
		args.scope = string(godot.TopLevelScope)
	}

	if !args.help && !args.version && len(args.files) == 0 {
		return arguments{}, fmt.Errorf("files list is empty")
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
