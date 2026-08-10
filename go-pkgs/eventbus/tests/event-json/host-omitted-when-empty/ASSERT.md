## Expected

- Marshaled JSON does **not** contain a `"host"` key.
- Other fields (`id`, `ts`, `source`, `type`, `payload`) are still present.
- Round-trip `Host` is empty string.

## Side Effects

- None.

## Errors

- `err` is nil.

## Exit Code

- Success.

```go
import (
	"encoding/json"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("event-json host omitempty: unexpected error: %v", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(resp.JSONBytes, &raw); err != nil {
		t.Fatalf("marshaled JSON invalid: %v\n%s", err, resp.JSONBytes)
	}
	if _, ok := raw["host"]; ok {
		t.Fatalf("expected host to be omitted when empty, got JSON: %s", resp.JSONBytes)
	}
	for _, key := range []string{"id", "ts", "source", "type", "payload"} {
		if _, ok := raw[key]; !ok {
			t.Fatalf("marshaled JSON missing key %q: %s", key, resp.JSONBytes)
		}
	}
	if resp.RoundTrip.Host != "" {
		t.Fatalf("RoundTrip.Host: got %q, want empty", resp.RoundTrip.Host)
	}
}
```
