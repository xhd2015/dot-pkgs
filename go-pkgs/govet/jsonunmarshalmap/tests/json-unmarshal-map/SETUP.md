## Preconditions
- The `jsonunmarshalmap` package is importable.
- Each leaf's `Setup` sets `req.Src` to the Go source to analyze.

## Steps
1. Parse `req.Src` with `go/parser`.
2. Run `jsonunmarshalmap.Checker.CheckAST`.
3. Return collected violations.

## Context
- `DOCTEST_ROOT` points to this test tree's root directory.

```go
import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/jsonunmarshalmap"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Request struct {
	Src string
}

type Response struct {
	Violations []types.Violation
}

func Run(t *testing.T, req *Request) (*Response, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "fixture.go", req.Src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse source: %w", err)
	}
	violations := (&jsonunmarshalmap.Checker{}).CheckAST(fset, f)
	return &Response{Violations: violations}, nil
}
```
