# Scenario

**Feature**: json-unmarshal-map AST checker

```
# parse fixture Go source -> jsonunmarshalmap.Checker -> violations
fixture.go -> CheckAST -> []Violation
```

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
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}
```