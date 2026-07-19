# Scenario

**Feature**: two complete CSI 6n queries in one buffer yield two concatenated CPRs

```
# multi-query
Data = ESC[6n + ESC[6n, Row=4, Col=8
  -> replies ESC[4;8R ESC[4;8R
```

## Steps

1. Set `req.Data` to two consecutive `\x1b[6n` sequences.
2. Set cursor to `(4,8)`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Data = []byte("\x1b[6n\x1b[6n")
	req.Row = 4
	req.Col = 8
	return nil
}
```
