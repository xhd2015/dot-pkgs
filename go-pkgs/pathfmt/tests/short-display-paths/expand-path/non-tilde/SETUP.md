# Scenario

**Feature**: non-tilde absolute paths are returned unchanged by `Expand`

```
# Expand passthrough
no ~ prefix -> unchanged
```

## Steps

1. Set `req.Path` to a platform absolute path without a `~` prefix.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Path = "/abs/path"
	return nil
}```
