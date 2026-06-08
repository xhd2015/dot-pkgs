package cleanfunc

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestChecker_NoFunctions(t *testing.T) {
	c := &Checker{MaxLines: 64}
	fset, f := parseAST(t, `package main
var x = 1
`)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Errorf("expected no violations for file with no functions, got %d", len(v))
	}
}

func TestChecker_WithinLimit(t *testing.T) {
	c := &Checker{MaxLines: 64}
	fset, f := parseAST(t, `package main
func foo() {
	// short function
	println("hello")
}
`)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Errorf("expected no violations for short function, got %d", len(v))
	}
}

func TestChecker_AtThreshold(t *testing.T) {
	c := &Checker{MaxLines: 64}
	body := strings.Repeat("x := 1\n", 62)
	src := "package main\nfunc foo() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Fatalf("expected no violations at threshold (64 lines), got %d", len(v))
	}
}

func TestChecker_ExceedsThreshold(t *testing.T) {
	c := &Checker{MaxLines: 64}
	body := strings.Repeat("x := 1\n", 63)
	src := "package main\nfunc foo() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for function exceeding 64 lines, got %d", len(v))
	}
	if v[0].Checker != "clean-func" {
		t.Errorf("expected checker 'clean-func', got %q", v[0].Checker)
	}
	if v[0].Line == 0 {
		t.Errorf("expected non-zero line number, got %d", v[0].Line)
	}
	if !strings.Contains(v[0].Message, "foo") {
		t.Errorf("expected message to contain function name, got %q", v[0].Message)
	}
}

func TestChecker_Method(t *testing.T) {
	c := &Checker{MaxLines: 64}
	body := strings.Repeat("x := 1\n", 63)
	src := "package main\ntype T struct{}\nfunc (t T) bar() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation for long method, got %d", len(v))
	}
	if !strings.Contains(v[0].Message, "bar") {
		t.Errorf("expected message to contain method name 'bar', got %q", v[0].Message)
	}
}

func TestChecker_MultipleFunctions(t *testing.T) {
	c := &Checker{MaxLines: 64}
	body := strings.Repeat("x := 1\n", 63)
	src := "package main\n" +
		"func short() { println(\"ok\") }\n" +
		"func long() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation (only long function), got %d", len(v))
	}
	if !strings.Contains(v[0].Message, "long") {
		t.Errorf("expected violation for 'long', got %q", v[0].Message)
	}
}

func TestChecker_TwoLongFunctions(t *testing.T) {
	c := &Checker{MaxLines: 64}
	body := strings.Repeat("x := 1\n", 63)
	src := "package main\n" +
		"func a() {\n" + body + "}\n" +
		"func b() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(v))
	}
}

func TestChecker_FunctionWithoutBody(t *testing.T) {
	c := &Checker{MaxLines: 64}
	fset, f := parseAST(t, `package main
func foo()
`)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Errorf("expected no violations for function declaration without body, got %d", len(v))
	}
}

func TestChecker_CustomMaxLines(t *testing.T) {
	c := &Checker{MaxLines: 10}
	body := strings.Repeat("x := 1\n", 8)
	src := "package main\nfunc foo() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 0 {
		t.Fatalf("expected no violations with custom max=10 and 10-line func, got %d", len(v))
	}

	body = strings.Repeat("x := 1\n", 9)
	src = "package main\nfunc foo() {\n" + body + "}\n"
	fset, f = parseAST(t, src)
	v = c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation with custom max=10 and 11-line func, got %d", len(v))
	}
}

func TestChecker_ZeroMaxLinesDefaults(t *testing.T) {
	c := &Checker{MaxLines: 0}
	body := strings.Repeat("x := 1\n", 63)
	src := "package main\nfunc foo() {\n" + body + "}\n"
	fset, f := parseAST(t, src)
	v := c.CheckAST(fset, f)
	if len(v) != 1 {
		t.Fatalf("expected 1 violation when MaxLines=0 (defaults to 64), got %d", len(v))
	}
}

func TestChecker_Name(t *testing.T) {
	c := &Checker{}
	if c.Name() != "clean-func" {
		t.Errorf("expected name 'clean-func', got %q", c.Name())
	}
}

func parseAST(t *testing.T, src string) (*token.FileSet, *ast.File) {
	t.Helper()
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return fset, f
}
