# Scenario

**Feature**: pure `MergeProcessEnv` without spawn TERM policy

```
# merge surface
req.ApplySpawnTERM = false
  -> Run calls only MergeProcessEnv(base, set, unset)
  -> no EnsureSpawnTERM
```

## Preconditions

- This branch asserts pure merge semantics only (identity, set, unset, order).

## Steps

1. Set `req.ApplySpawnTERM` to false so TERM policy is not applied.

## Context

- Leaves supply concrete `Base` / `Set` / `Unset`.
- Empty set and unset mean those slices are nil or empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ApplySpawnTERM = false
	return nil
}
```
