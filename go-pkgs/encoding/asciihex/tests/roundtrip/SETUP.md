# Scenario

**Feature**: Decode inverts Encode for the same byte slice

```
# composition — raw 0xff stays one byte
caller Data -> Encode -> Hex -> Decode -> Data
```

## Preconditions

- `req.Op` is `"roundtrip"`.
- Leaves set `req.Data`. `Run` encodes then decodes that slice.

## Steps

1. Set `req.Op` to `"roundtrip"`.
2. `Run` records both `resp.Encoded` and `resp.Decoded`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "roundtrip"
	return nil
}
```
