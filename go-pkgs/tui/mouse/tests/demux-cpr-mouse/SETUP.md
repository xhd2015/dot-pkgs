# Scenario

**Feature**: DemuxCPR peels CPR and forwards SGR mouse sequences

```
# byte stream: CPR then SGR mouse press
ESC[41;1R + ESC[<0;67;25M
  -> events=[{41,1}], forward=mouse bytes, rest empty
```

## Preconditions

- `req.Op = "demux"`.
- Hold is empty; data is one CPR followed by one complete SGR mouse.

## Steps

1. Set Op demux and DemuxData to CPR + mouse SGR bytes.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "demux"
	req.DemuxHold = nil
	req.DemuxData = []byte("\x1b[41;1R\x1b[<0;67;25M")
	return nil
}
```
