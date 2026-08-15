# Scenario

**Feature**: err-unchecked-or-ignored AST checker

```
# parse fixture Go source -> errcheck.Checker -> violations
fixture.go -> CheckAST -> []Violation
```

## Preconditions

- The `errcheck` package is importable.
- Each leaf's `Setup` sets `req.Src` to the Go source to analyze.

## Steps

1. Parse `req.Src` with `go/parser`.
2. Run `errcheck.Checker.CheckAST`.
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