# Scenario

**Feature**: pure `ParseLsofFn` extracts unique absolute paths from `lsof -Fn`

```
lsof -Fn fixture bytes -> ParseLsofFn -> []string absolute paths (unique)
```

## Preconditions

- Leaves supply `req.LsofOutput`.
- No live `lsof` invocation.

## Steps

1. Set `req.Op` to `"parse-lsof"`.
2. Leaf fills `req.LsofOutput`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "parse-lsof"
	return nil
}
```
