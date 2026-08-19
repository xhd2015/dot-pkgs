## Expected

- `resp.OK` is true.
- `joinSpans(resp.Spans)` equals `"brainstorm"`.
- Spans are `b` (matched), `rain` (unmatched), `s` (matched), `tor` (unmatched), `m` (matched).

## Errors

- `err` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/fuzzy"
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
	want := []fuzzy.Span{
		{Text: "b", Matched: true},
		{Text: "rain", Matched: false},
		{Text: "s", Matched: true},
		{Text: "tor", Matched: false},
		{Text: "m", Matched: true},
	}
	if len(resp.Spans) != len(want) {
		t.Fatalf("spans = %#v, want %#v", resp.Spans, want)
	}
	for i := range want {
		if resp.Spans[i].Text != want[i].Text || resp.Spans[i].Matched != want[i].Matched {
			t.Fatalf("spans = %#v, want %#v", resp.Spans, want)
		}
	}
}
```
