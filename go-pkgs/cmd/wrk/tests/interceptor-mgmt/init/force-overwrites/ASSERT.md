## Expected

- Exit code 0.
- Empty stdout preferred.
- Interceptor replaced with disabled neutral stub (not `custom-tool`).

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
	assertMgmtNeutralStub(t, req.WrkHome)
	ic := readMgmtInterceptor(t, req.WrkHome)
	for _, a := range ic.Argv {
		if a == "custom-tool" || a == "replace-me" {
			t.Fatalf("force init should replace custom argv, still has %q in %v", a, ic.Argv)
		}
	}
}
```
