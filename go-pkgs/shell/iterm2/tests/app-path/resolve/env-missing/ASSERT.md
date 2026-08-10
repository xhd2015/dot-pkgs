## Expected

- Resolved path is `""` (not home, not system, not the missing env path).
- Documents localbot-consistent strictness: env set but unusable → no fallthrough.

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
	if resp.ResolvedPath != "" {
		t.Fatalf("env set but missing must not fall through: got %q (home=%q system=%q env=%q)",
			resp.ResolvedPath, homeApp(req.HomeDir), systemApp, req.EnvValue)
	}
}
```
