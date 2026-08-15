# Scenario

**Feature**: `ListOwned` error paths before and during gh execution

```
# validation or gh failure
ListOwned -> validate Options OR exec gh -> error returned to caller
```

## Preconditions

- Error leaves expect `ListOwned` to return a non-nil error.
- Validation leaves must not invoke mock `gh`.

## Steps

1. Descendant `Setup` configures `req` to trigger the specific error.
2. `Run` returns `(nil, err)` or partial response with error.

## Context

- Gh error leaves use mock scripts that exit non-zero or print invalid JSON.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if req.Limit == 0 {
		req.Limit = 100
	}
	return nil
}```