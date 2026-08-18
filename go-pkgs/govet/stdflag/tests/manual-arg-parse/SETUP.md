# Scenario

**Feature**: manual-arg-parse AST checker

```
# parse fixture Go source -> stdflag checkers -> violations
fixture.go -> CheckAST -> []Violation
```

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
- Each leaf's `Setup` populates `req.Src` by reading `fixture.go` via `fixtureFile` (`DOCTEST_CASE`).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	return nil
}

func fixtureFile(d *session.Doctest, rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	base := d.DOCTEST_CASE
	if base == "" || !filepath.IsAbs(base) {
		base = filepath.Join(d.DOCTEST_ROOT, base)
	}
	return filepath.Join(base, rel)
}
```