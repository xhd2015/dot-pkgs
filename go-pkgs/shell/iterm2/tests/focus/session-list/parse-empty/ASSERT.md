## Expected

- Empty string → empty slice, nil error (via Run: `resp.Refs` len 0).
- Whitespace-only and comment-only inputs → empty slice, nil error.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Refs) != 0 {
		t.Fatalf("Run parse empty: len(refs)=%d, want 0; refs=%+v", len(resp.Refs), resp.Refs)
	}
	inputs := []string{
		req.ListOutput,
		"",
		"   \n\t\n",
		"# comment only\n# another\n",
	}
	for _, in := range inputs {
		refs, perr := iterm2.ParseSessionListOutput(in)
		if perr != nil {
			t.Fatalf("ParseSessionListOutput(%q) error: %v", in, perr)
		}
		if len(refs) != 0 {
			t.Fatalf("ParseSessionListOutput(%q) len=%d, want 0; refs=%+v", in, len(refs), refs)
		}
	}
}
```
