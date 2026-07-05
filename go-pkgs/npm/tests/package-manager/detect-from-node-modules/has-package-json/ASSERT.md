## Expected

- `resp.Trace.HasPackageJSON` is true.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Trace.HasPackageJSON {
		t.Fatal("expected HasPackageJSON = true")
	}
}```
