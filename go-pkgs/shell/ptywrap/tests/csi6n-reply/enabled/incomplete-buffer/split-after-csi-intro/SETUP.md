# Scenario

**Feature**: split after `ESC[` still completes on second chunk

```
# split after CSI intro
chunk1 = ESC[
chunk2 = 6n
  -> final write ESC[3;7R
```

## Steps

1. Set `req.Chunks` to `{\x1b[}`, then `{6n}`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Chunks = [][]byte{
		[]byte{0x1b, '['},
		[]byte("6n"),
	}
	return nil
}
```
