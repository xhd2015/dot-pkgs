# Scenario

**Feature**: empty set and unset leave base environ unchanged

```
# identity
base + empty set + empty unset
  -> MergeProcessEnv
  -> same keys/values as base
```

## Preconditions

- Set and Unset are empty; only Base is meaningful.

## Steps

1. Leave `req.Set` and `req.Unset` empty (leaf fills Base).

## Context

- Map equality of KEY=value pairs is the contract; order may match base.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Set = nil
	req.Unset = nil
	return nil
}
```
