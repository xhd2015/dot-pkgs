## Expected

- Inner quotes are backslash-escaped for AppleScript literals.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := `/tmp/\"proj\"`
	if resp.Escaped != want {
		t.Fatalf("Escaped = %q, want %q", resp.Escaped, want)
	}
}
```