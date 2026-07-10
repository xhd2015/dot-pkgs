## Expected Output

Pretty JSON for the interceptor object (indent + trailing newline), e.g.:

```
{
  "enabled": true,
  "argv": [
    "kool",
    "space",
    "create"
  ]
}
```

## Expected

- Exit code 0.
- Stdout is valid pretty JSON of `create.interceptor` with `enabled: true` and the seeded argv.
- Trailing newline after the last content line.

## Exit Code

- 0

```go
import (
	"encoding/json"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("stdout must end with newline: %q", resp.Stdout)
	}
	var ic mgmtInterceptorSpec
	if err := json.Unmarshal([]byte(resp.Stdout), &ic); err != nil {
		t.Fatalf("stdout not JSON interceptor: %v\n%s", err, resp.Stdout)
	}
	if !ic.Enabled {
		t.Fatalf("enabled: want true, got false")
	}
	wantArgv := []string{"kool", "space", "create"}
	if len(ic.Argv) != len(wantArgv) {
		t.Fatalf("argv len: want %d got %d (%v)", len(wantArgv), len(ic.Argv), ic.Argv)
	}
	for i := range wantArgv {
		if ic.Argv[i] != wantArgv[i] {
			t.Fatalf("argv[%d]: want %q got %q", i, wantArgv[i], ic.Argv[i])
		}
	}
	// Prefer pretty indent (multi-line) for human readability.
	if !strings.Contains(resp.Stdout, "\n") || strings.Count(resp.Stdout, "\n") < 2 {
		t.Fatalf("expected pretty multi-line JSON, got %q", resp.Stdout)
	}
}
```
