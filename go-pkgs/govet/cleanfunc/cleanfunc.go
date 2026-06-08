package cleanfunc

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct {
	MaxLines int
}

func (c *Checker) Name() string {
	return "clean-func"
}

func (c *Checker) CheckAST(fset *token.FileSet, file *ast.File) []types.Violation {
	maxLines := c.MaxLines
	if maxLines <= 0 {
		maxLines = 64
	}

	var violations []types.Violation
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		if fn.Body == nil {
			continue
		}
		startPos := fset.Position(fn.Pos())
		endPos := fset.Position(fn.End())
		lines := endPos.Line - startPos.Line + 1
		if lines > maxLines {
			violations = append(violations, types.Violation{
				File:    startPos.Filename,
				Line:    startPos.Line,
				Col:     startPos.Column,
				Message: fmt.Sprintf("function %s has %d lines, exceeds maximum of %d", fn.Name.Name, lines, maxLines),
				Checker: c.Name(),
			})
		}
	}
	return violations
}
