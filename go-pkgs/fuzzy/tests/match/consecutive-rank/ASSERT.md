## Expected

- `resp.OK` is true (`Match("ab", "ab")`).
- `Match("a-b", "ab")` is also OK.
- `resp.Score` is strictly greater than the gapped match score.

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
		t.Fatal("expected Match(\"ab\", \"ab\") OK")
	}
	gapped, gerr := Run(t, d, &Request{Op: "match", Haystack: "a-b", Query: "ab"})
	if gerr != nil {
		t.Fatal(gerr)
	}
	if !gapped.OK {
		t.Fatal("expected Match(\"a-b\", \"ab\") OK")
	}
	if resp.Score <= gapped.Score {
		t.Fatalf("consecutive Match(\"ab\",\"ab\") score %d should exceed gapped Match(\"a-b\",\"ab\") score %d", resp.Score, gapped.Score)
	}
}
```
