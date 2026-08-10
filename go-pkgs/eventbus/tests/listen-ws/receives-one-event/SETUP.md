# Scenario

**Feature**: ListenWS receives one Event from a test WebSocket server

```
# one-frame smoke
WS server -> text JSON Event -> ListenWS onEvent -> cancel after first
```

## Steps

1. Start a WS mock that writes the fixture Event once.
2. Set `req.WSURL` to the `ws://` form of the mock URL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	srv := startWSEventServer(t, req.Event)
	req.WSURL = httpToWSURL(srv.URL)
	return nil
}
```
