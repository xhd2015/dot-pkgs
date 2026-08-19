## Expected

- `resp.OK` is true.
- `joinSpans(resp.Spans)` equals the haystack.
- `"aid"` appears as a matched span (original case, not `"AID"`).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.OK {
		t.Fatalf("expected Match(%q, %q) OK", req.Haystack, req.Query)
	}
	if joinSpans(resp.Spans) != req.Haystack {
		t.Fatalf("joinSpans = %q, want haystack %q", joinSpans(resp.Spans), req.Haystack)
	}
	got := matchedTexts(resp.Spans)
	foundAid := false
	for _, s := range got {
		if s == "AID" {
			t.Fatalf("matched Text must keep haystack case, got %q in %#v", s, got)
		}
		if s == "aid" {
			foundAid = true
		}
	}
	if !foundAid {
		t.Fatalf("expected matched span %q, got %#v", "aid", got)
	}
}
```
