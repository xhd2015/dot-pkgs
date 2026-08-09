# Scenario

**Feature**: unset then set reintroduces KEY with the set value

```
# operation order
base -> remove Unset keys -> apply Set
  -> KEY from set present even if also listed in unset
```

## Preconditions

- Product order is locked: Unset first, Set second.

## Steps

1. Keep merge surface (`ApplySpawnTERM=false`).
2. Leaves combine Unset and Set for the same key.

## Context

- This is the S5 contract path used when callers clear then replace a variable.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	// Parent merge/ already cleared TERM policy; re-assert for this interaction branch.
	req.ApplySpawnTERM = false
	return nil
}
```
