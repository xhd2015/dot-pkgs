## Expected

- `resp.OK` is true.
- `joinSpans(resp.Spans)` equals the haystack.
- `"aid"` and `"user"` appear as matched span texts.

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
		t.Fatalf("expected MatchAll(%q, Tokens(%q)) OK", req.Haystack, req.Query)
	}
	if joinSpans(resp.Spans) != req.Haystack {
		t.Fatalf("joinSpans = %q, want haystack %q", joinSpans(resp.Spans), req.Haystack)
	}
	got := matchedTexts(resp.Spans)
	hasAid, hasUser := false, false
	for _, s := range got {
		if s == "aid" {
			hasAid = true
		}
		if s == "user" {
			hasUser = true
		}
	}
	if !hasAid || !hasUser {
		t.Fatalf("expected matched spans %q and %q, got %#v", "aid", "user", got)
	}
}
```
