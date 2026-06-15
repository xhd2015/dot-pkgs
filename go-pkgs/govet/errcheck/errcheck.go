package errcheck

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct{}

func (c *Checker) Name() string {
	return "err-unchecked-or-ignored"
}

func (c *Checker) CheckAST(fset *token.FileSet, file *ast.File) []types.Violation {
	var violations []types.Violation

	errorVars := collectErrorVars(file)

	ast.Inspect(file, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok {
			return true
		}
		ident, ok := isErrNilCheck(ifStmt.Cond)
		if !ok {
			return true
		}
		if !errorVars[ident.Name] && !hasErrName(ident.Name) {
			return true
		}
		returns := collectReturns(ifStmt.Body)
		for _, ret := range returns {
			if !identAppearsInReturn(ident, ret) {
				pos := fset.Position(ifStmt.Pos())
				violations = append(violations, types.Violation{
					File:    pos.Filename,
					Line:    pos.Line,
					Col:     pos.Column,
					Checker: c.Name(),
					Message: "error checked but not propagated; returns fallback value instead of error",
					Hint:    "propagate the error: return zero-value, err or wrap it with fmt.Errorf; if the function signature lacks an error return, consider adding one",
				})
			}
		}
		return true
	})

	return violations
}

func collectErrorVars(file *ast.File) map[string]bool {
	errorVars := make(map[string]bool)
	ast.Inspect(file, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		if assign.Tok != token.DEFINE && assign.Tok != token.ASSIGN {
			return true
		}
		if len(assign.Lhs) < 2 {
			return true
		}
		lastIdx := len(assign.Lhs) - 1
		ident, ok := assign.Lhs[lastIdx].(*ast.Ident)
		if !ok {
			return true
		}
		errorVars[ident.Name] = true
		return true
	})
	return errorVars
}

func hasErrName(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "err")
}

func isErrNilCheck(cond ast.Expr) (*ast.Ident, bool) {
	binExpr, ok := cond.(*ast.BinaryExpr)
	if !ok {
		return nil, false
	}
	if binExpr.Op != token.NEQ {
		return nil, false
	}

	xIdent, xOk := binExpr.X.(*ast.Ident)
	yIdent, yOk := binExpr.Y.(*ast.Ident)

	if xOk && yOk && yIdent.Name == "nil" && xIdent.Name != "nil" {
		return xIdent, true
	}
	if xOk && yOk && xIdent.Name == "nil" && yIdent.Name != "nil" {
		return yIdent, true
	}
	return nil, false
}

func collectReturns(body *ast.BlockStmt) []*ast.ReturnStmt {
	var returns []*ast.ReturnStmt
	ast.Inspect(body, func(n ast.Node) bool {
		ret, ok := n.(*ast.ReturnStmt)
		if ok {
			returns = append(returns, ret)
		}
		return true
	})
	return returns
}

func identAppearsInReturn(ident *ast.Ident, ret *ast.ReturnStmt) bool {
	for _, expr := range ret.Results {
		if identInExpr(ident, expr) {
			return true
		}
	}
	return false
}

func identInExpr(want *ast.Ident, expr ast.Expr) bool {
	found := false
	ast.Inspect(expr, func(n ast.Node) bool {
		if found {
			return false
		}
		if id, ok := n.(*ast.Ident); ok && id.Name == want.Name {
			found = true
			return false
		}
		return true
	})
	return found
}
