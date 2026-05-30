package dupstat

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractFunctions(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "sample.go")
	src := `package test

func Hello() {
	fmt.Println("hello")
}

type Service struct{}

func (s *Service) Greet(name string) string {
	if name == "" {
		return "hi"
	}
	return "hello " + name
}

func NoBody() // no body, should be skipped
`
	if err := os.WriteFile(filePath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	funcs, err := ExtractFunctions(filePath, "pkg/test")
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 2 {
		t.Fatalf("expected 2 functions (one with body, one with NoBody skipped), got %d", len(funcs))
	}

	helloFunc := funcs[0]
	if helloFunc.Name != "Hello" {
		t.Errorf("expected Hello, got %s", helloFunc.Name)
	}
	if helloFunc.Receiver != "" {
		t.Errorf("expected no receiver, got %q", helloFunc.Receiver)
	}
	if helloFunc.PkgPath != "pkg/test" {
		t.Errorf("expected pkg/test, got %s", helloFunc.PkgPath)
	}
	if !strings.Contains(string(helloFunc.BodySrc), "fmt.Println") {
		t.Errorf("expected body to contain fmt.Println, got %s", string(helloFunc.BodySrc))
	}

	greetFunc := funcs[1]
	if greetFunc.Name != "Greet" {
		t.Errorf("expected Greet, got %s", greetFunc.Name)
	}
	if greetFunc.Receiver != "*Service" {
		t.Errorf("expected receiver *Service, got %q", greetFunc.Receiver)
	}
}

func TestExtractFunctionsNoBodies(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "empty.go")
	src := `package test

func NoBody()

type T struct{}
func (t T) Method()
`
	if err := os.WriteFile(filePath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	funcs, err := ExtractFunctions(filePath, "pkg/test")
	if err != nil {
		t.Fatal(err)
	}

	if len(funcs) != 0 {
		t.Fatalf("expected 0 functions (all no body), got %d", len(funcs))
	}
}

func TestPackagePath(t *testing.T) {
	result := PackagePath("/root/module", "/root/module/pkg/auth/login.go")
	if result != "pkg/auth" {
		t.Errorf("expected pkg/auth, got %s", result)
	}
}

func TestTypeExprToString(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "types.go")

	t.Run("*Service", func(t *testing.T) {
		src := "package p\ntype Service struct{}\nfunc (t *Service) M() { return }\n"
		if err := os.WriteFile(filePath, []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		funcs, err := ExtractFunctions(filePath, "p")
		if err != nil {
			t.Fatal(err)
		}
		if len(funcs) == 0 {
			t.Fatal("expected one function")
		}
		if funcs[0].Receiver != "*Service" {
			t.Errorf("got %q, want %q", funcs[0].Receiver, "*Service")
		}
	})

	t.Run("Service", func(t *testing.T) {
		src := "package p\ntype Service struct{}\nfunc (t Service) M() { return }\n"
		if err := os.WriteFile(filePath+".2", []byte(src), 0644); err != nil {
			t.Fatal(err)
		}
		funcs, err := ExtractFunctions(filePath+".2", "p")
		if err != nil {
			t.Fatal(err)
		}
		if len(funcs) == 0 {
			t.Fatal("expected one function")
		}
		if funcs[0].Receiver != "Service" {
			t.Errorf("got %q, want %q", funcs[0].Receiver, "Service")
		}
	})
}

func TestTokenizeFunction(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "f.go")
	src := "package p\nfunc Add(a, b int) int {\n\treturn a + b\n}\n"
	if err := os.WriteFile(filePath, []byte(src), 0644); err != nil {
		t.Fatal(err)
	}

	funcs, err := ExtractFunctions(filePath, "p")
	if err != nil || len(funcs) == 0 {
		t.Fatal("expected one function")
	}

	tokens := TokenizeFunction(funcs[0])
	if len(tokens.Raw) == 0 {
		t.Error("expected raw tokens")
	}
	if len(tokens.Norm) == 0 {
		t.Error("expected normalized tokens")
	}
	if len(tokens.Mixed) == 0 {
		t.Error("expected mixed tokens")
	}
	if tokens.Func == nil {
		t.Error("expected func reference")
	}
}
