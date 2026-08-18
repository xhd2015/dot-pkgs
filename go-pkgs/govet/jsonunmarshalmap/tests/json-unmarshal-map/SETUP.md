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
- Leaves read `fixture.go` via `fixtureFile` (`DOCTEST_CASE`), not process cwd.

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