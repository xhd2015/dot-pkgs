# Scenario

**Feature**: Unset removes keys from base before set is applied

```
# unset branch
base + Unset keys
  -> MergeProcessEnv
  -> listed keys absent; others preserved
```

## Preconditions

- Set is empty on this branch (no reintroduction).

## Steps

1. Clear `req.Set` by default; leaves set Base and Unset.

## Context

- Unsetting a key not present in base is a no-op (absent-key-noop leaf).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Set = nil
	return nil
}
```
