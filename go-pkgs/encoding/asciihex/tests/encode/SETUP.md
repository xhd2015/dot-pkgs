# Scenario

**Feature**: Encode maps each input byte to a lowercase `\xHH` step

```
# library string — no trailing newline
caller Data -> Encode -> concatenated \xHH
```

## Preconditions

- `req.Op` is `"encode"`.
- Leaves set `req.Data`. `Run` calls `asciihex.Encode(req.Data)` only.

## Steps

1. Set `req.Op` to `"encode"`.
2. `Run` records `resp.Encoded`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "encode"
	return nil
}
```
