## Expected

- `NormalizeTTY("ttys148")` equals `NormalizeTTY("/dev/ttys148")`.
- Both non-empty results are non-empty strings.
- `NormalizeTTY("")` is empty.
- `NormalizeTTY` result for `/dev/ttys148` matches `resp.Normalized` from Run.

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
	bare := iterm2.NormalizeTTY("ttys148")
	dev := iterm2.NormalizeTTY("/dev/ttys148")
	if bare == "" || dev == "" {
		t.Fatalf("NormalizeTTY bare/dev must be non-empty; bare=%q dev=%q", bare, dev)
	}
	if bare != dev {
		t.Fatalf("NormalizeTTY(\"ttys148\")=%q != NormalizeTTY(\"/dev/ttys148\")=%q", bare, dev)
	}
	if empty := iterm2.NormalizeTTY(""); empty != "" {
		t.Fatalf("NormalizeTTY(\"\") = %q, want empty", empty)
	}
	if resp.Normalized != dev {
		t.Fatalf("resp.Normalized = %q, want %q (Run path)", resp.Normalized, dev)
	}
	// Additional pair for coverage of another TTY number.
	if a, b := iterm2.NormalizeTTY("ttys001"), iterm2.NormalizeTTY("/dev/ttys001"); a != b || a == "" {
		t.Fatalf("NormalizeTTY ttys001 pair: bare=%q dev=%q", a, b)
	}
}
```
