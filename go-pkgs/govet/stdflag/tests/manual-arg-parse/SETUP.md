## Preconditions
- The `stdflag` package is importable from the test module.
- Each leaf directory contains a `fixture.go` file with the Go source to analyze.

## Steps
1. Read the Go source from `req.Src`.
2. Parse it with `go/parser` in `ParseComments` mode.
3. Run both `stdflag.Checker` (import detection) and `stdflag.ManualFlagChecker` (manual parsing).
4. Return all collected violations.

## Context
- `DOCTEST_ROOT` points to this test tree's root directory.
- Each leaf's `Setup` populates `req.Src` by reading `fixture.go` from the leaf directory.

```go
import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/stdflag"
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
	var violations []types.Violation
	violations = append(violations, (&stdflag.Checker{}).CheckAST(fset, f)...)
	violations = append(violations, (stdflag.ManualFlagChecker{}).CheckAST(fset, f)...)
	return &Response{Violations: violations}, nil
}
```
