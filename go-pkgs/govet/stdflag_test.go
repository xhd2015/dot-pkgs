package govet

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func parseAST(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return fset, f
}

func TestStdFlagChecker_NoImports(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main`)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Errorf("expected no violations for no imports, got %d", len(v))
	}
}

func TestStdFlagChecker_NonFlagImport(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main
import "fmt"`)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Errorf("expected no violations for fmt import, got %d", len(v))
	}
}

func TestStdFlagChecker_StdFlagImport(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main
import "flag"`)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(v))
	}
	if v[0].Checker != "std-flag" {
		t.Errorf("expected checker 'std-flag', got %q", v[0].Checker)
	}
	if !strings.Contains(v[0].Message, "flag") {
		t.Errorf("expected message about 'flag', got %q", v[0].Message)
	}
	if !strings.Contains(v[0].Message, "less-flags") {
		t.Errorf("expected message to suggest 'less-flags', got %q", v[0].Message)
	}
	if !strings.Contains(v[0].Hint, "flags-parsing") {
		t.Errorf("expected hint about flags-parsing, got %q", v[0].Hint)
	}
}

func TestStdFlagChecker_AliasedFlagImport(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main
import f "flag"`)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for aliased flag import, got %d", len(v))
	}
}

func TestStdFlagChecker_LessFlagsImport(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main
import "github.com/xhd2015/less-flags"`)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Errorf("expected no violations for less-flags import, got %d", len(v))
	}
}

func TestStdFlagChecker_FlagAndLessFlags(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main
import (
	"flag"
	"github.com/xhd2015/less-flags"
)`)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for flag + less-flags, got %d", len(v))
	}
}

func TestStdFlagChecker_MultipleFlagImports(t *testing.T) {
	c := &StdFlagChecker{}
	fset, f := parseAST(t, `package main
import (
	"flag"
	f "flag"
)`)
	v := c.CheckAST(fset, f)
	if len(v) != 2 {
		t.Fatalf("expected 2 violations for two flag imports, got %d", len(v))
	}
}

func TestStdFlagChecker_Name(t *testing.T) {
	c := &StdFlagChecker{}
	if c.Name() != "std-flag" {
		t.Errorf("expected name 'std-flag', got %q", c.Name())
	}
}
