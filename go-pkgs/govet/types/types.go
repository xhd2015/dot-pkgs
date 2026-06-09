package types

import (
	"go/ast"
	"go/token"
)

type Violation struct {
	File    string
	Line    int
	Col     int
	Message string
	Checker string
	Hint    string
}

type Config struct {
	FileMaxLines    int
	ExcludeCheckers []string
	Files           []string
}

type FileChecker interface {
	Name() string
	CheckFile(filename string, src []byte) []Violation
}

type ASTChecker interface {
	Name() string
	CheckAST(fset *token.FileSet, file *ast.File) []Violation
}
