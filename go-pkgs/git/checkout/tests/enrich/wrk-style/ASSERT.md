## Expected

- `resp.Meta.Error` is empty.
- `resp.Meta.Branch` is `"main"`.
- `resp.Meta.CommitSHA` has length 7.
- `resp.Meta.CommitMsg` is `"wrk fixture"`.
- `resp.Meta.Status` is `"clean"`.

## Errors

- `err` from `Run` is nil (Enrich never returns error to caller).

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	if m.CommitMsg != "wrk fixture" {
		t.Fatalf("CommitMsg = %q, want wrk fixture", m.CommitMsg)
	}
	if m.Status != "clean" {
		t.Fatalf("Status = %q, want clean", m.Status)
	}
}
```