## Expected

- At least one event was delivered to the ListenWS callback.
- The first received event matches `req.Event` on ID, Type, Source, Host, and TS.
- `err` is nil **or** is a context cancellation/deadline error after the event was received
  (Run cancels ctx on first event).

## Side Effects

- One WebSocket dial and at least one text frame read.

## Errors

- Accept nil or context.Canceled / context.DeadlineExceeded when events were received.
- Other errors fail the test.

## Exit Code

- Success when an event was received.

```go
import (
	"context"
	"errors"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if len(resp.ReceivedEvents) == 0 {
		t.Fatalf("ListenWS: expected at least one event, err=%v", err)
	}
	if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("ListenWS: unexpected error after receiving events: %v", err)
	}
	got := resp.ReceivedEvents[0]
	if got.ID != req.Event.ID {
		t.Fatalf("ID: got %q, want %q", got.ID, req.Event.ID)
	}
	if got.Type != req.Event.Type {
		t.Fatalf("Type: got %q, want %q", got.Type, req.Event.Type)
	}
	if got.Source != req.Event.Source {
		t.Fatalf("Source: got %q, want %q", got.Source, req.Event.Source)
	}
	if got.Host != req.Event.Host {
		t.Fatalf("Host: got %q, want %q", got.Host, req.Event.Host)
	}
	if got.TS != req.Event.TS {
		t.Fatalf("TS: got %q, want %q", got.TS, req.Event.TS)
	}
}
```
