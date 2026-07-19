# Scenario

**Feature**: split before final `n` still completes on second chunk

```
# split before final n
chunk1 = ESC[6
chunk2 = n
  -> final write ESC[3;7R
```

## Steps

1. Set `req.Chunks` to `{\x1b[6}`, then `{n}`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Chunks = [][]byte{
		[]byte{0x1b, '[', '6'},
		[]byte("n"),
	}
	return nil
}
```
