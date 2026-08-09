# Scenario

**Feature**: Set assignments apply after base (last-wins)

```
# set branch
base + Set KEY=value entries
  -> MergeProcessEnv
  -> KEY present with final set value
```

## Preconditions

- Unset is empty on this branch unless a leaf overrides.

## Steps

1. Clear `req.Unset` by default; leaves set Base and Set.

## Context

- Duplicate keys within Set: last entry wins (S3).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Unset = nil
	return nil
}
```
