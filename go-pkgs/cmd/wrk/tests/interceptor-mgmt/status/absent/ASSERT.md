## Expected Output

```
---
version: 2
---
state: absent
path: {WRK_HOME}/config.json
argv0: -
```

## Expected

- Exit code 0.
- Three-line status with `state: absent` and `argv0: -`.

## Exit Code

- 0

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, statusStdoutTemplate("absent", mgmtConfigPath(req.WrkHome), "-"))
}
```
