package jsonunmarshalmap

import (
	"go/ast"
	"go/token"
	"strconv"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Checker struct{}

func (c *Checker) Name() string {
	return "json-unmarshal-map"
}

func (c *Checker) CheckAST(fset *token.FileSet, file *ast.File) []types.Violation {
	jsonName := ""
	for _, imp := range file.Imports {
		path, err := strconv.Unquote(imp.Path.Value)
		if err != nil {
			continue
		}
		if path == "encoding/json" {
			if imp.Name != nil {
				jsonName = imp.Name.Name
			} else {
				jsonName = "json"
			}
			break
		}
	}
	if jsonName == "" {
		return nil
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

		typeMap := make(map[string]bool)
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch stmt := n.(type) {
			case *ast.DeclStmt:
				genDecl, ok := stmt.Decl.(*ast.GenDecl)
				if !ok || genDecl.Tok != token.VAR {
					break
				}
				for _, spec := range genDecl.Specs {
					vs, ok := spec.(*ast.ValueSpec)
					if !ok {
						continue
					}
					if !isStringAnyMap(vs.Type) {
						continue
					}
					for _, name := range vs.Names {
						typeMap[name.Name] = true
					}
				}
			case *ast.AssignStmt:
				if stmt.Tok != token.DEFINE {
					break
				}
				for i, rhs := range stmt.Rhs {
					typ := getMapTypeFromRhs(rhs)
					if typ == nil || !isStringAnyMap(typ) {
						continue
					}
					if i < len(stmt.Lhs) {
						if ident, ok := stmt.Lhs[i].(*ast.Ident); ok {
							typeMap[ident.Name] = true
						}
					}
				}
			}
			return true
		})

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			pkgIdent, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if pkgIdent.Name != jsonName || sel.Sel.Name != "Unmarshal" {
				return true
			}
			if len(call.Args) < 2 {
				return true
			}
			arg := call.Args[1]
			unary, ok := arg.(*ast.UnaryExpr)
			if !ok || unary.Op != token.AND {
				return true
			}
			if ident, ok := unary.X.(*ast.Ident); ok {
				if typeMap[ident.Name] {
					pos := fset.Position(call.Pos())
					violations = append(violations, types.Violation{
						File:    pos.Filename,
						Line:    pos.Line,
						Col:     pos.Column,
						Message: "json.Unmarshal target is map[string]any or map[string]interface{}; consider using a typed struct instead",
						Checker: c.Name(),
					})
				}
				return true
			}
			if compLit, ok := unary.X.(*ast.CompositeLit); ok {
				if isStringAnyMap(compLit.Type) {
					pos := fset.Position(call.Pos())
					violations = append(violations, types.Violation{
						File:    pos.Filename,
						Line:    pos.Line,
						Col:     pos.Column,
						Message: "json.Unmarshal target is map[string]any or map[string]interface{}; consider using a typed struct instead",
						Checker: c.Name(),
					})
				}
			}
			return true
		})
	}

	return violations
}

func isStringAnyMap(expr ast.Expr) bool {
	if expr == nil {
		return false
	}
	mt, ok := expr.(*ast.MapType)
	if !ok {
		return false
	}
	keyIdent, ok := mt.Key.(*ast.Ident)
	if !ok || keyIdent.Name != "string" {
		return false
	}
	return isAnyOrInterface(mt.Value)
}

func isAnyOrInterface(expr ast.Expr) bool {
	if ident, ok := expr.(*ast.Ident); ok && ident.Name == "any" {
		return true
	}
	if iface, ok := expr.(*ast.InterfaceType); ok {
		return iface.Methods == nil || len(iface.Methods.List) == 0
	}
	return false
}

func getMapTypeFromRhs(expr ast.Expr) ast.Expr {
	switch e := expr.(type) {
	case *ast.CompositeLit:
		return e.Type
	case *ast.CallExpr:
		if ident, ok := e.Fun.(*ast.Ident); ok && ident.Name == "make" && len(e.Args) > 0 {
			return e.Args[0]
		}
	}
	return nil
}
