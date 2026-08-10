## Expected

- `ParseTabSetFindOutput` returns two `TabSessionRef` values, nil error.
- Both have `SetName == "bots"`.
- Tab IDs are `t1` and `t2` (order preserved).
- TTY / window fields populated from fixture when present.

## Exit Code

- N/A (parse-find phase)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {

	if err != nil {
		t.Fatal(err)
	}
	refs, perr := iterm2.ParseTabSetFindOutput(req.FindOutput)
	if perr != nil {
		t.Fatalf("ParseTabSetFindOutput error: %v", perr)
	}
	if len(refs) != 2 {
		t.Fatalf("len(refs) = %d, want 2; refs=%+v", len(refs), refs)
	}
	if refs[0].SetName != "bots" || refs[1].SetName != "bots" {
		t.Fatalf("SetName want bots for both; got %q, %q", refs[0].SetName, refs[1].SetName)
	}
	if refs[0].TabID != "t1" {
		t.Fatalf("refs[0].TabID = %q, want t1", refs[0].TabID)
	}
	if refs[1].TabID != "t2" {
		t.Fatalf("refs[1].TabID = %q, want t2", refs[1].TabID)
	}
	if refs[0].TTY != "/dev/ttys001" {
		t.Fatalf("refs[0].TTY = %q, want /dev/ttys001", refs[0].TTY)
	}
	if refs[1].TTY != "/dev/ttys002" {
		t.Fatalf("refs[1].TTY = %q, want /dev/ttys002", refs[1].TTY)
	}
	if refs[0].WindowID != "win-A" {
		t.Fatalf("refs[0].WindowID = %q, want win-A", refs[0].WindowID)
	}
}
```
