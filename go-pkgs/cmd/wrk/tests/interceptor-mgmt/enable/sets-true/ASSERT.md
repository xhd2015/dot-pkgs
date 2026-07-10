## Expected

- Exit code 0.
- Empty stdout.
- Config has `enabled: true`; argv unchanged.

## Exit Code

- 0

```go
import (
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmptyStdout(t, resp.Stdout)
	assertMgmtInterceptorEnabled(t, req.WrkHome, true)
	ic := readMgmtInterceptor(t, req.WrkHome)
	if len(ic.Argv) < 1 || ic.Argv[0] != "echo" {
		t.Fatalf("argv should be preserved, got %v", ic.Argv)
	}
}
```
