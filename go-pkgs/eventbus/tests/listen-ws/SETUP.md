# Scenario

**Feature**: ListenWS dials a WebSocket and decodes JSON Event frames

```
# thin listen helper
ListenWS(ctx, wsURL, onEvent) -> read JSON frames until ctx cancel
```

## Steps

1. Set `req.Op` to `"listen-ws"`.
2. Leaves start a WS mock and set `req.WSURL`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "listen-ws"
	return nil
}
```
