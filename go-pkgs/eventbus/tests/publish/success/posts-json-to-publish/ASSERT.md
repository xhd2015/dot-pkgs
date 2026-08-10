## Expected

- `Publish` returns nil.
- Exactly one HTTP request was captured.
- Method is `POST`.
- Path is `/publish`.
- `Content-Type` indicates JSON (`application/json`, optionally with charset).
- Body unmarshals to an Event matching `req.Event` fields (id, ts, source, type, host).
- `Authorization` header is empty (token not set).

## Side Effects

- One POST to the mock hub.

## Errors

- `err` is nil.

## Exit Code

- Success.

```go
import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/eventbus"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("Publish: unexpected error: %v", err)
	}
	if req.Capture.Len() != 1 {
		t.Fatalf("expected exactly 1 HTTP request, got %d", req.Capture.Len())
	}
	got, ok := req.Capture.Last()
	if !ok {
		t.Fatal("missing captured request")
	}
	if got.Method != "POST" {
		t.Fatalf("Method: got %q, want POST", got.Method)
	}
	if got.Path != "/publish" {
		t.Fatalf("Path: got %q, want /publish", got.Path)
	}
	if !strings.HasPrefix(got.ContentType, "application/json") {
		t.Fatalf("Content-Type: got %q, want application/json…", got.ContentType)
	}
	if got.Authorization != "" {
		t.Fatalf("Authorization: got %q, want empty when token empty", got.Authorization)
	}
	var ev eventbus.Event
	if err := json.Unmarshal(got.Body, &ev); err != nil {
		t.Fatalf("body JSON: %v\n%s", err, got.Body)
	}
	if ev.ID != req.Event.ID || ev.TS != req.Event.TS || ev.Source != req.Event.Source ||
		ev.Type != req.Event.Type || ev.Host != req.Event.Host {
		t.Fatalf("body Event mismatch:\n got %+v\nwant %+v", ev, req.Event)
	}
}
```
