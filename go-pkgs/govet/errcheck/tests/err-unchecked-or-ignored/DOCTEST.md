# Error Unchecked or Ignored Detection Tests

## Version
0.0.2

Tests for the `err-unchecked-or-ignored` checker that detects error variables
that are checked (`if err != nil`) but not propagated in the return statement
(returning a fallback/default value instead of the error).

# DSN (Domain Specific Notion)

- **Checker** — `errcheck.Checker` analyzes Go AST for mishandled errors.
- **Fixture source** — each leaf provides `req.Src` Go source to parse.
- **Violations** — collected `types.Violation` records from the checker.

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
	"github.com/xhd2015/dot-pkgs/go-pkgs/govet/errcheck"
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
	violations := (&errcheck.Checker{}).CheckAST(fset, f)
	return &Response{Violations: violations}, nil
}
```