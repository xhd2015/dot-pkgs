## Expected

- `resp.Meta.Error` is `"no commits (HEAD unborn)"`.
- Branch, sha, msg, and status fields are empty.

## Errors

- `err` from `Run` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	m := resp.Meta
	if m.Error != "no commits (HEAD unborn)" {
		t.Fatalf("Error = %q, want no commits (HEAD unborn)", m.Error)
	}
	if m.Branch != "" || m.CommitSHA != "" || m.CommitMsg != "" || m.Status != "" {
		t.Fatalf("expected empty fields on unborn HEAD, got %+v", m)
	}
}
```
