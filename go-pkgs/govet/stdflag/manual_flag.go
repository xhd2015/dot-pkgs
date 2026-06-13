package stdflag

import (
	"go/ast"
	"go/token"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type ManualFlagChecker struct{}

func (c ManualFlagChecker) Name() string {
	return "manual-flag-parse"
}

func (c ManualFlagChecker) CheckAST(fset *token.FileSet, file *ast.File) []types.Violation {
	var violations []types.Violation

	ast.Inspect(file, func(n ast.Node) bool {
		switch stmt := n.(type) {
		case *ast.ForStmt:
			if stmt.Body != nil {
				inspectForBody(stmt.Body, fset, &violations)
			}
			return false
		case *ast.RangeStmt:
			if stmt.Body != nil {
				inspectForBody(stmt.Body, fset, &violations)
			}
			return false
		}
		return true
	})

	return violations
}

func inspectForBody(body *ast.BlockStmt, fset *token.FileSet, violations *[]types.Violation) {
	ast.Inspect(body, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.ForStmt:
			if n.Body != nil {
				inspectForBody(n.Body, fset, violations)
			}
			return false
		case *ast.RangeStmt:
			if n.Body != nil {
				inspectForBody(n.Body, fset, violations)
			}
			return false
		case *ast.SwitchStmt:
			if hasFlagPrefixCase(n) {
				pos := fset.Position(n.Pos())
				*violations = append(*violations, types.Violation{
					File:    pos.Filename,
					Line:    pos.Line,
					Col:     pos.Column,
					Message: "manual argument parsing detected; consider using 'github.com/xhd2015/less-flags' instead",
					Checker: "manual-flag-parse",
					Hint:    "run `go-best-practice flags-parsing` for guidance",
				})
			}
		case *ast.IfStmt:
			if hasFlagPrefixComparison(n) {
				pos := fset.Position(n.Pos())
				*violations = append(*violations, types.Violation{
					File:    pos.Filename,
					Line:    pos.Line,
					Col:     pos.Column,
					Message: "manual argument parsing detected; consider using 'github.com/xhd2015/less-flags' instead",
					Checker: "manual-flag-parse",
					Hint:    "run `go-best-practice flags-parsing` for guidance",
				})
			}
		}
		return true
	})
}

func hasFlagPrefixCase(switchStmt *ast.SwitchStmt) bool {
	for _, stmt := range switchStmt.Body.List {
		caseClause, ok := stmt.(*ast.CaseClause)
		if !ok {
			continue
		}
		for _, expr := range caseClause.List {
			lit, ok := expr.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				continue
			}
			if isFlagString(lit.Value) {
				return true
			}
		}
	}
	return false
}

func hasFlagPrefixComparison(ifStmt *ast.IfStmt) bool {
	binExpr, ok := ifStmt.Cond.(*ast.BinaryExpr)
	if !ok {
		return false
	}
	if binExpr.Op != token.EQL && binExpr.Op != token.NEQ {
		return false
	}
	return isFlagStringLit(binExpr.X) || isFlagStringLit(binExpr.Y)
}

func isFlagStringLit(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	return isFlagString(lit.Value)
}

func isFlagString(s string) bool {
	if len(s) < 3 {
		return false
	}
	return s[1] == '-'
}
