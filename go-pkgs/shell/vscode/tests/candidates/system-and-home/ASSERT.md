## Expected

- No error.
- Four candidates: VS Code and Insiders under `/Applications`, then the same
  two under `<Home>/Applications`.
- Every candidate ends with `Contents/Resources/app/bin/code`.
- The system VS Code path is present.

## Side Effects

- None: the function only builds paths.

## Errors

- None.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if got := len(resp.Candidates); got != 4 {
		t.Fatalf("candidates = %#v, want 4", resp.Candidates)
	}
	if n := countSuffix(t, resp.Candidates, "/Contents/Resources/app/bin/code"); n != 4 {
		t.Fatalf("candidates = %#v, want every path to end in the code CLI", resp.Candidates)
	}
	if !hasPrefix(resp.Candidates, req.Home+"/Applications/") {
		t.Fatalf("candidates = %#v, want a per-user home candidate under %s", resp.Candidates, req.Home)
	}
	want := systemCodePath("Visual Studio Code.app")
	found := false
	for _, p := range resp.Candidates {
		if p == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("candidates = %#v, want %s", resp.Candidates, want)
	}
}
```
