# Scenario

**Feature**: Decode walks a `\xHH` string into raw bytes or a kool-shaped error

```
# whole string must be 4-character \xHH steps
caller Hex -> Decode -> bytes
caller bad Hex -> Decode -> error
```

## Preconditions

- `req.Op` is `"decode"`.
- Leaves set `req.Hex`. `Run` calls `asciihex.Decode(req.Hex)`.

## Steps

1. Set `req.Op` to `"decode"`.
2. `Run` records `resp.Decoded` and the returned error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "decode"
	return nil
}
```
