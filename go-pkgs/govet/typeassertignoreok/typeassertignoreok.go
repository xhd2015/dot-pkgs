package typeassertignoreok

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct{}

func (c *Checker) Name() string {
	return "type-assert-ignore-ok"
}

func (c *Checker) CheckAST(fset *token.FileSet, file *ast.File) []types.Violation {
	var violations []types.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if hasTypeAssertWithIgnoredOk(assign) {
			pos := fset.Position(assign.Pos())
			violations = append(violations, types.Violation{
				File:    pos.Filename,
				Line:    pos.Line,
				Col:     pos.Column,
				Message: "type assertion ignores ok value; use the comma-ok pattern to handle type assertion failures safely",
				Checker: c.Name(),
			})
		}
		return true
	})

	return violations
}

func hasTypeAssertWithIgnoredOk(stmt *ast.AssignStmt) bool {
	if len(stmt.Lhs) != 2 {
		return false
	}
	if len(stmt.Rhs) != 1 {
		return false
	}
	if _, ok := stmt.Rhs[0].(*ast.TypeAssertExpr); !ok {
		return false
	}
	for _, lhs := range stmt.Lhs {
		if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "_" {
			return true
		}
	}
	return false
}
