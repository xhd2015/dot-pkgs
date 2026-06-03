package govet

import (
	"go/ast"
	"go/token"
	"strconv"
)

type StdFlagChecker struct{}

func (c *StdFlagChecker) Name() string {
	return "std-flag"
}

func (c *StdFlagChecker) CheckAST(fset *token.FileSet, file *ast.File) []Violation {
	var violations []Violation
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if path == "flag" {
			pos := fset.Position(imp.Pos())
			violations = append(violations, Violation{
				File:    pos.Filename,
				Line:    pos.Line,
				Col:     pos.Column,
				Message: "import of stdlib 'flag' detected; consider using 'github.com/xhd2015/less-flags' instead",
				Checker: c.Name(),
				Hint:    "run `go-best-practice flags-parsing` for guidance",
			})
		}
	}
	return violations
}
