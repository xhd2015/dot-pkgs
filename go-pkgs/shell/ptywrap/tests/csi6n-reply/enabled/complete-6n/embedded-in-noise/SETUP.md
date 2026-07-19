# Scenario

**Feature**: CSI 6n embedded in surrounding output still yields one CPR

```
# embedded
Data = "noise" + ESC[6n + "more"
  -> replies ESC[2;9R ; rest empty
```

## Steps

1. Set `req.Data` to noise-wrapped `\x1b[6n`.
2. Set cursor to `(2,9)`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Data = []byte("noise\x1b[6nmore")
	req.Row = 2
	req.Col = 9
	return nil
}
```
