package dupstat

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
)

type Function struct {
	Name     string
	Receiver string
	PkgPath  string
	File     string
	Line     int
	BodySrc  []byte
	SigSrc   []byte
}

type FunctionTokens struct {
	Func *Function
	Raw  []string
	Norm []string
	Mixed []string
}

func ExtractFunctions(filePath string, pkgPath string) ([]Function, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var funcs []Function

	ast.Inspect(f, func(n ast.Node) bool {
		fn, ok := n.(*ast.FuncDecl)
		if !ok {
			return true
		}

		if fn.Body == nil {
			return true
		}

		lbraceOff := fset.Position(fn.Body.Lbrace).Offset
		rbraceOff := fset.Position(fn.Body.Rbrace).Offset

		funcStartOff := fset.Position(fn.Type.Func).Offset

		bodySrc := src[lbraceOff : rbraceOff+1]
		sigSrc := src[funcStartOff:lbraceOff]

		var receiver string
		if fn.Recv != nil && len(fn.Recv.List) > 0 {
			receiver = typeExprToString(fn.Recv.List[0].Type)
		}

		funcs = append(funcs, Function{
			Name:     fn.Name.Name,
			Receiver: receiver,
			PkgPath:  pkgPath,
			File:     filePath,
			Line:     fset.Position(fn.Pos()).Line,
			BodySrc:  bodySrc,
			SigSrc:   sigSrc,
		})

		return true
	})

	return funcs, nil
}

func typeExprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + typeExprToString(t.X)
	case *ast.SelectorExpr:
		return typeExprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + typeExprToString(t.Elt)
		}
		return "[...]" + typeExprToString(t.Elt)
	default:
		return ""
	}
}

func TokenizeFunction(fn Function) FunctionTokens {
	raw := tokenizeRaw(fn.BodySrc)
	norm := tokenizeNormalized(fn.BodySrc)
	mixed := tokenizeMixed(fn.SigSrc, fn.BodySrc)
	return FunctionTokens{
		Func:  &fn,
		Raw:   raw,
		Norm:  norm,
		Mixed: mixed,
	}
}

func PackagePath(rootDir, filePath string) string {
	rel, err := filepath.Rel(rootDir, filePath)
	if err != nil {
		return ""
	}
	dir := filepath.Dir(rel)
	if dir == "." {
		return ""
	}
	return dir
}
