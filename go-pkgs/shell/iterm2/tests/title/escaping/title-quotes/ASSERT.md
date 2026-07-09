## Expected

- Backslash and double-quote are escaped for AppleScript string literals.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := `say \"hi\"\\x`
	if resp.Escaped != want {
		t.Fatalf("Escaped = %q, want %q", resp.Escaped, want)
	}
}
```
