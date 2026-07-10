## Expected

- Exit code 0.
- Empty stdout preferred.
- `config.json` exists with `create.interceptor.enabled: false` and non-empty argv (neutral stub).

## Side Effects

- Creates `{WRK_HOME}/config.json` with versioned structure.

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
}
```
