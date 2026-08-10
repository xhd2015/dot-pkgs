## Expected

- Resolved path equals the usable `ITERM2_APP_PATH` custom path.
- Must not return home or system candidates when env wins.

## Exit Code

- N/A (library)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := req.EnvValue
	if resp.ResolvedPath != want {
		t.Fatalf("ResolveAppPathWith() = %q, want env path %q", resp.ResolvedPath, want)
	}
	if resp.ResolvedPath == homeApp(req.HomeDir) || resp.ResolvedPath == systemApp {
		t.Fatalf("env must win over home/system; got %q", resp.ResolvedPath)
	}
}
```
