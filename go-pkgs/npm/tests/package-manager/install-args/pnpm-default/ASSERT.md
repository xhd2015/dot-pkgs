## Expected

- `resp.Args` is `["install"]`.
- `resp.Command` is `"pnpm"`.

## Errors

- `err` is nil.

```go
import (
	"reflect"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"install"}
	if !reflect.DeepEqual(resp.Args, want) {
		t.Fatalf("InstallArgs = %v, want %v", resp.Args, want)
	}
	if resp.Command != "pnpm" {
		t.Fatalf("command = %q, want %q", resp.Command, "pnpm")
	}
}
```