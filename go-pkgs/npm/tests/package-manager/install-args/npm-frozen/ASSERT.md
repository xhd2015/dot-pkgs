## Expected

- `resp.Args` is `["ci"]`.
- `resp.Command` is `"npm"`.

## Errors

- `err` is nil.

```go
import (
	"reflect"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"ci"}
	if !reflect.DeepEqual(resp.Args, want) {
		t.Fatalf("InstallArgs = %v, want %v", resp.Args, want)
	}
	if resp.Command != "npm" {
		t.Fatalf("command = %q, want %q", resp.Command, "npm")
	}
}
```