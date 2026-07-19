# Scenario

**Feature**: CSI CUP `ESC[H` is not a cursor query — no CPR

```
# non-match CUP
Data = ESC[H
  -> replies empty; rest empty
```

## Steps

1. Set `req.Data` to `\x1b[H`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Data = []byte("\x1b[H")
	return nil
}
```
