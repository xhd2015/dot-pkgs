## Expected

- `Scan` returns no error.
- Exactly 3 modules, sorted by `Dir` lexical: `.`, `app`, `nested/service`.
- `Dir == "."` maps to `Path == "example.com/root"`.
- `Dir == "app"` maps to `Path == "example.com/root/app"`.
- `Dir == "nested/service"` maps to `Path == "example.com/root/service"` (slash-joined,
  no `./` prefix, `filepath.Rel` style).

```go
import (
	"reflect"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Scan(%q) failed: %v", req.RootDir, resp.Err)
	}

	gotDirs := dirLines(resp.Modules)
	wantDirs := []string{".", "app", "nested/service"}
	if !reflect.DeepEqual(gotDirs, wantDirs) {
		t.Fatalf("Scan dirs = %v, want %v (sorted by Dir)", gotDirs, wantDirs)
	}

	if p := pathOf(resp.Modules, "."); p != "example.com/root" {
		t.Fatalf("Path for '.' = %q, want example.com/root", p)
	}
	if p := pathOf(resp.Modules, "app"); p != "example.com/root/app" {
		t.Fatalf("Path for 'app' = %q, want example.com/root/app", p)
	}
	if p := pathOf(resp.Modules, "nested/service"); p != "example.com/root/service" {
		t.Fatalf("Path for 'nested/service' = %q, want example.com/root/service", p)
	}
}
```
