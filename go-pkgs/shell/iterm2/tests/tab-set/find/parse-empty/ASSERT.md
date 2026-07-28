## Expected

- Empty string → empty slice, nil error.
- Whitespace-only and comment-only lines → empty slice, nil error.

## Exit Code

- N/A (parse-find phase)

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
	inputs := []string{
		req.FindOutput,
		"",
		"   \n\t\n",
		"# comment only\n# another\n",
	}
	for _, in := range inputs {
		refs, perr := iterm2.ParseTabSetFindOutput(in)
		if perr != nil {
			t.Fatalf("ParseTabSetFindOutput(%q) error: %v", in, perr)
		}
		if len(refs) != 0 {
			t.Fatalf("ParseTabSetFindOutput(%q) len=%d, want 0; refs=%+v", in, len(refs), refs)
		}
	}
}
```
