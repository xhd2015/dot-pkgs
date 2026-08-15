# Manual Argument Parsing Detection Tests

## Version
0.0.2

Tests for the `manual-flag-parse` checker that detects manual flag parsing patterns
(`for`/`switch`, `for`/`if`) in Go source code and suggests using a proper flag library.

# DSN (Domain Specific Notion)

- **Checker** — `stdflag.Checker` and `stdflag.ManualFlagChecker` analyze Go AST.
- **Fixture source** — each leaf provides `req.Src` Go source to parse.
- **Violations** — collected `types.Violation` records from both checkers.

## How to Run

```sh
doctest test -v ./
```

```go
import (
	"fmt"
	"go/parser"
	"go/token"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/stdflag"
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/types"
)

type Request struct {
	Src string
}

type Response struct {
	Violations []types.Violation
}

func Run(t *testing.T, d *session.Doctest, req *Request) (*Response, error) {
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