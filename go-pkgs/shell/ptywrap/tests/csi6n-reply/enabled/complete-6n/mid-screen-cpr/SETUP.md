# Scenario

**Feature**: complete `ESC[6n` with mid-screen cursor replies `ESC[5;12R`

```
# mid-screen CPR
Data=ESC[6n, Row=5, Col=12
  -> replies ESC[5;12R
```

## Steps

1. Set `req.Data` to `\x1b[6n`.
2. Set cursor to 1-based `(5,12)` (as if `cursor.Y=4`, `cursor.X=11`).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Data = []byte("\x1b[6n")
	req.Row = 5
	req.Col = 12
	return nil
}
```
