# Scenario

**Feature**: List / ListLinked inventory via go-pkgs package path

```
# leaf builds main (+ linked)
go-pkgs worktree.List / ListLinked -> Entry slices
```

## Preconditions

- Root helpers for init / linked worktrees.

## Steps

1. Grouping only; leaves set `req.Op` and fixture paths.

## Context

- Proves go-pkgs still exposes list after shim to gitops.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	// Grouping: leaves set Op and build fixtures.
	return nil
}
```
