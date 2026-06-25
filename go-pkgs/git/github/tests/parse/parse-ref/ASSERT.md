## Expected

- `resp.ParseRefOwner` is `xhd2015`.
- `resp.ParseRefName` is `fixture-a`.

## Errors

- `err` is nil.

```go
import "testing"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ParseRefOwner != "xhd2015" || resp.ParseRefName != "fixture-a" {
		t.Fatalf("ParseRef = %q/%q, want xhd2015/fixture-a", resp.ParseRefOwner, resp.ParseRefName)
	}
}
```