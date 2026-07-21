# Scenario

**Feature**: type-assert-ignore-ok AST checker

```
# parse fixture Go source -> typeassertignoreok.Checker -> violations
fixture.go -> CheckAST -> []Violation
```

## Preconditions

- The `typeassertignoreok` package is importable.
- Each leaf's `Setup` sets `req.Src` to the Go source to analyze.

## Steps

1. Parse `req.Src` with `go/parser`.
2. Run `typeassertignoreok.Checker.CheckAST`.
3. Return collected violations.

## Context

- `d.DOCTEST_ROOT` points to this test tree's root directory.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	return nil
}
```