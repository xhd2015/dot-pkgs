## Expected

- `resp.OK` is true.
- `resp.Score` is `0`.
- `resp.Spans` is one unmatched span whose `Text` is `"foo"`.
- `joinSpans(resp.Spans)` equals the haystack.

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
		t.Fatal("expected Match(\"foo\", \"\") OK")
	}
	if resp.Score != 0 {
		t.Fatalf("expected score 0, got %d", resp.Score)
	}
	if joinSpans(resp.Spans) != "foo" {
		t.Fatalf("joinSpans = %q, want haystack %q", joinSpans(resp.Spans), "foo")
	}
	want := []fuzzy.Span{{Text: "foo", Matched: false}}
	if len(resp.Spans) != 1 || resp.Spans[0].Text != want[0].Text || resp.Spans[0].Matched {
		t.Fatalf("expected one unmatched span {foo,false}, got %#v", resp.Spans)
	}
}
```
