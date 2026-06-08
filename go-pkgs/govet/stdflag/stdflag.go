package stdflag

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct{}

func (c *Checker) Name() string {
	return "std-flag"
}

func (c *Checker) CheckAST(fset *token.FileSet, file *ast.File) []types.Violation {
	var violations []types.Violation
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if path == "flag" {
			pos := fset.Position(imp.Pos())
			violations = append(violations, types.Violation{
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
