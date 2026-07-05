## Expected

- `resp.Args` is `["install", "--frozen-lockfile"]`.
- `resp.Command` is `"yarn"`.

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
	want := []string{"install", "--frozen-lockfile"}
	if !reflect.DeepEqual(resp.Args, want) {
		t.Fatalf("InstallArgs = %v, want %v", resp.Args, want)
	}
	if resp.Command != "yarn" {
		t.Fatalf("command = %q, want %q", resp.Command, "yarn")
	}
}
```