## Expected

- `Scan` returns non-nil error (abort after `mmm-mid`).
- `OnEntryOrder` is `["aaa-first", "mmm-mid"]` — one callback per completed entry
  in sorted order before abort.
- `zzz-last` never appears in callback order or full results.

## Errors

- Nil error when abort expected.
- Callback order not sorted or includes post-abort entry.

```go
import (
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err == nil {
		t.Fatal("expected abort error from OnEntry, got nil")
	}
	if !strings.Contains(err.Error(), "abort after mmm-mid") {
		t.Fatalf("error = %v, want abort after mmm-mid", err)
	}

	wantOrder := []string{"aaa-first", "mmm-mid"}
	if len(resp.OnEntryOrder) != len(wantOrder) {
		t.Fatalf("OnEntryOrder = %v, want %v", resp.OnEntryOrder, wantOrder)
	}
	for i, name := range wantOrder {
		if resp.OnEntryOrder[i] != name {
			t.Fatalf("OnEntryOrder[%d] = %q, want %q; full %v", i, resp.OnEntryOrder[i], name, resp.OnEntryOrder)
		}
	}

	for _, e := range resp.Entries {
		if e.Name == "zzz-last" {
			t.Fatalf("zzz-last should not be scanned after abort; entries = %d", len(resp.Entries))
		}
	}
}
```