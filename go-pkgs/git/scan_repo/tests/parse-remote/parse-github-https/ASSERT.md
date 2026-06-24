## Expected

- `resp.ParseOK` is true.
- `resp.Owner` is `"golang"`.
- `resp.Repo` is `"go"`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.ParseOK {
		t.Fatal("expected ParseOK true for GitHub HTTPS URL")
	}
	if resp.Owner != "golang" {
		t.Fatalf("owner = %q, want golang", resp.Owner)
	}
	if resp.Repo != "go" {
		t.Fatalf("repo = %q, want go", resp.Repo)
	}
}
```