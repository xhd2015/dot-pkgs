# Scenario

**Feature**: Grok account usage Fetch (auth ensure → HTTP billing → Snapshot)

```
# L2 harness (parallel-safe)
leaf Setup -> Operation + FetchMode on Request
root Run    -> usage.Fetch with injectable Ensure/GetJSON
leaf Assert -> Snapshot fields / error
```

## Preconditions

- Package under test:
  `github.com/xhd2015/dot-pkgs/go-pkgs/shell/grok/usage`
- Injected auth only (no real auth.json). Fixture email `user@example.com`.
- No real network.

## Steps

1. Root Setup clears Operation/FetchMode.
2. Grouping Setup sets Operation.
3. Leaf Setup sets FetchMode.
4. Root Run calls Fetch.
5. Leaf Assert checks Snapshot or error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = ""
	req.FetchMode = ""
	return nil
}
```
