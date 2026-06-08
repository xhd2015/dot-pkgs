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
	FileMaxLines int
	Excludes     []string
	Paths        []string
}

type FileChecker interface {
	Name() string
	CheckFile(filename string, src []byte) []Violation
}

type ASTChecker interface {
	Name() string
	CheckAST(fset *token.FileSet, file *ast.File) []Violation
}
