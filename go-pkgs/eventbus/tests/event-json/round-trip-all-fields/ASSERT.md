## Expected

- `resp.JSONBytes` is valid JSON containing keys: `id`, `ts`, `source`, `type`, `host`, `payload`.
- `resp.RoundTrip` equals `req.Event` for ID, TS, Source, Type, Host.
- Payload JSON is semantically equal to the original payload object.

## Side Effects

- None.

## Errors

- `err` is nil.

## Exit Code

- Success.

```go
import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("event-json round-trip: unexpected error: %v", err)
	}
	if len(resp.JSONBytes) == 0 {
		t.Fatal("expected non-empty marshaled JSON")
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resp.JSONBytes, &raw); err != nil {
		t.Fatalf("marshaled JSON invalid: %v\n%s", err, resp.JSONBytes)
	}
	for _, key := range []string{"id", "ts", "source", "type", "host", "payload"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("marshaled JSON missing key %q: %s", key, resp.JSONBytes)
		}
	}
	rt := resp.RoundTrip
	if rt.ID != req.Event.ID {
		t.Fatalf("ID: got %q, want %q", rt.ID, req.Event.ID)
	}
	if rt.TS != req.Event.TS {
		t.Fatalf("TS: got %q, want %q", rt.TS, req.Event.TS)
	}
	if rt.Source != req.Event.Source {
		t.Fatalf("Source: got %q, want %q", rt.Source, req.Event.Source)
	}
	if rt.Type != req.Event.Type {
		t.Fatalf("Type: got %q, want %q", rt.Type, req.Event.Type)
	}
	if rt.Host != req.Event.Host {
		t.Fatalf("Host: got %q, want %q", rt.Host, req.Event.Host)
	}
	if !bytes.Equal(bytes.TrimSpace(rt.Payload), bytes.TrimSpace(req.Event.Payload)) {
		// compare as JSON values to tolerate spacing
		var a, b any
		if err := json.Unmarshal(rt.Payload, &a); err != nil {
			t.Fatalf("round-trip payload: %v", err)
		}
		if err := json.Unmarshal(req.Event.Payload, &b); err != nil {
			t.Fatalf("request payload: %v", err)
		}
		ab, _ := json.Marshal(a)
		bb, _ := json.Marshal(b)
		if !bytes.Equal(ab, bb) {
			t.Fatalf("Payload: got %s, want %s", rt.Payload, req.Event.Payload)
		}
	}
}
```
