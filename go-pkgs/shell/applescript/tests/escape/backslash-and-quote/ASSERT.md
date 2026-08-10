## Expected

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
	// \ → \\ , " → \"
	want := `echo \"hi\"\\`
	if resp.Escaped != want {
		t.Fatalf("Escaped = %q, want %q", resp.Escaped, want)
	}
}
```

