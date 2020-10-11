package godot

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"strings"
)

// getComments extracts comments from a file.
func getComments(file *ast.File, fset *token.FileSet, scope Scope) ([]comment, error) {
	var comments []comment

	// Render AST representation to a string
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return nil, fmt.Errorf("render file: %v", err)
	}
	lines := strings.Split(buf.String(), "\n")

	// All comments
	if scope == AllScope {
		cc, err := getAllComments(file, fset, lines)
		if err != nil {
			return nil, fmt.Errorf("get all comments: %v", err)
		}
		return append(comments, cc...), nil
	}

	// Comments from the inside of top level blocks
	cc, err := getBlockComments(file, fset, lines)
	if err != nil {
		return nil, fmt.Errorf("get block comments: %v", err)
	}
	comments = append(comments, cc...)

	// All top level comments
	if scope == TopLevelScope {
		cc, err := getTopLevelComments(file, fset, lines)
		if err != nil {
			return nil, fmt.Errorf("get top level comments: %v", err)
		}
		return append(comments, cc...), nil
	}

	// Top level declaration comments
	cc, err = getDeclarationComments(file, fset, lines)
	if err != nil {
		return nil, fmt.Errorf("get declaration comments: %v", err)
	}
	comments = append(comments, cc...)

	return comments, nil
}

// getBlockComments gets comments from the inside of top level
// blocks: var (...), const (...).
func getBlockComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment
	for _, decl := range file.Decls {
		d, ok := decl.(*ast.GenDecl)
		if !ok {
			continue
		}
		// No parenthesis == no block
		if d.Lparen == 0 {
			continue
		}
		for _, c := range file.Comments {
			// Skip comments outside this block
			if d.Lparen > c.Pos() || c.Pos() > d.Rparen {
				continue
			}
			// Skip comments that are not top-level for this block
			// (the block itself is top level, so comments inside this block
			// would be on column 2)
			// nolint: gomnd
			if fset.Position(c.Pos()).Column != 2 {
				continue
			}
			firstLine := fset.Position(c.Pos()).Line
			lastLine := fset.Position(c.End()).Line
			if lastLine >= len(lines) {
				return nil, fmt.Errorf(
					"invalid line number inside comment: %s:%d",
					fset.Position(c.Pos()).Filename,
					fset.Position(c.Pos()).Line,
				)
			}
			comments = append(comments, comment{
				ast:   c,
				lines: lines[firstLine-1 : lastLine],
			})
		}
	}
	return comments, nil
}

// getTopLevelComments gets all top level comments.
func getTopLevelComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment // nolint: prealloc
	for _, c := range file.Comments {
		if fset.Position(c.Pos()).Column != 1 {
			continue
		}
		firstLine := fset.Position(c.Pos()).Line
		lastLine := fset.Position(c.End()).Line
		if lastLine >= len(lines) {
			return nil, fmt.Errorf(
				"invalid line number inside comment: %s:%d",
				fset.Position(c.Pos()).Filename,
				fset.Position(c.Pos()).Line,
			)
		}
		comments = append(comments, comment{
			ast:   c,
			lines: lines[firstLine-1 : lastLine],
		})
	}
	return comments, nil
}

// getDeclarationComments gets top level declaration comments.
func getDeclarationComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment
	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.GenDecl:
			if d.Doc != nil {
				firstLine := fset.Position(d.Doc.Pos()).Line
				lastLine := fset.Position(d.Doc.End()).Line
				if lastLine >= len(lines) {
					return nil, fmt.Errorf(
						"invalid line number inside comment: %s:%d",
						fset.Position(d.Doc.Pos()).Filename,
						fset.Position(d.Doc.Pos()).Line,
					)
				}
				comments = append(comments, comment{
					ast:   d.Doc,
					lines: lines[firstLine-1 : lastLine],
				})
			}
		case *ast.FuncDecl:
			if d.Doc != nil {
				firstLine := fset.Position(d.Doc.Pos()).Line
				lastLine := fset.Position(d.Doc.End()).Line
				if lastLine >= len(lines) {
					return nil, fmt.Errorf(
						"invalid line number inside comment: %s:%d",
						fset.Position(d.Doc.Pos()).Filename,
						fset.Position(d.Doc.Pos()).Line,
					)
				}
				comments = append(comments, comment{
					ast:   d.Doc,
					lines: lines[firstLine-1 : lastLine],
				})
			}
		}
	}
	return comments, nil
}

// getAllComments gets every single comment from the file.
func getAllComments(file *ast.File, fset *token.FileSet, lines []string) ([]comment, error) {
	var comments []comment //nolint: prealloc
	for _, c := range file.Comments {
		firstLine := fset.Position(c.Pos()).Line
		lastLine := fset.Position(c.End()).Line
		if lastLine >= len(lines) {
			return nil, fmt.Errorf(
				"invalid line number inside comment: %s:%d",
				fset.Position(c.Pos()).Filename,
				fset.Position(c.Pos()).Line,
			)
		}
		comments = append(comments, comment{
			ast:   c,
			lines: lines[firstLine-1 : lastLine],
		})
	}
	return comments, nil
}

// getText extracts text from comment. If comment is a special block
// (e.g., CGO code), a block of empty lines is returned. If comment contains
// special lines (e.g., tags or indented code examples), they are replaced
// with an empty line. The result can be multiline.
func getText(comment *ast.CommentGroup) (s string) {
	if len(comment.List) == 1 &&
		strings.HasPrefix(comment.List[0].Text, "/*") &&
		isSpecialBlock(comment.List[0].Text) {
		return ""
	}

	for _, c := range comment.List {
		text := c.Text
		isBlock := false
		if strings.HasPrefix(c.Text, "/*") {
			isBlock = true
			text = strings.TrimPrefix(text, "/*")
			text = strings.TrimSuffix(text, "*/")
		}
		for _, line := range strings.Split(text, "\n") {
			if isSpecialLine(line) {
				s += "\n"
				continue
			}
			if !isBlock {
				line = strings.TrimPrefix(line, "//")
			}
			s += line + "\n"
		}
	}
	if len(s) == 0 {
		return ""
	}
	return s[:len(s)-1] // trim last "\n"
}
