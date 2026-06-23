# Scenario

**Feature**: empty input is returned unchanged by `Expand`

```
# Expand passthrough
empty -> unchanged
```

## Steps

1. Set `req.Path` to `""`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Path = ""
	return nil
}```
