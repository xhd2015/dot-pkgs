## Expected

- `resp.Meta.Error` is empty.
- `resp.Meta.Branch` is `"main"`.
- `resp.Meta.CommitSHA` has length 7.
- `resp.Meta.CommitMsg` is `"wip"`.
- `resp.Meta.Status` is `"dirty (1 modified)"`.

## Errors

- `err` from `Run` is nil.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	m := resp.Meta
	if m.Error != "" {
		t.Fatalf("Error = %q, want empty", m.Error)
	}
	if m.Branch != "main" {
		t.Fatalf("Branch = %q, want main", m.Branch)
	}
	if len(m.CommitSHA) != 7 {
		t.Fatalf("CommitSHA = %q, want 7-char short sha", m.CommitSHA)
	}
	if m.CommitMsg != "wip" {
		t.Fatalf("CommitMsg = %q, want wip", m.CommitMsg)
	}
	if m.Status != "dirty (1 modified)" {
		t.Fatalf("Status = %q, want dirty (1 modified)", m.Status)
	}
}
```
