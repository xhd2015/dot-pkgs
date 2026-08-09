# Scenario

**Feature**: `Expand` converts `~`-prefixed display paths back to absolute paths

```
# Expand pipeline
display path string -> Expand -> ~ rules -> absolute path

# passthrough
empty or no ~ prefix -> unchanged
```

## Preconditions

- Expand tests do not depend on cwd; only `req.Path` matters.

## Steps

1. Set `req.Op` to `"expand"` for all leaves in this group.

## Context

- Leaves set `req.Path` to tilde, non-tilde, or empty inputs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "expand"
	return nil
}```
