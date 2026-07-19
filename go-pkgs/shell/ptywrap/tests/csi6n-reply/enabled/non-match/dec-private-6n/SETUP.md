# Scenario

**Feature**: DEC private `ESC[?6n` is not plain CSI 6n — no CPR

```
# non-match DEC private
Data = ESC[?6n
  -> replies empty; rest empty
```

## Steps

1. Set `req.Data` to `\x1b[?6n`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Data = []byte("\x1b[?6n")
	return nil
}
```
