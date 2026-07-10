## Expected Output

```
---
version: 2
---
{WRK_HOME}/config.json
```

## Expected

- Exit code 0.
- Stdout is the absolute path of `{WRK_HOME}/config.json` plus trailing `\n`.
- File may still be missing.

## Side Effects

- Read-only; no `config.json` created.

## Exit Code

- 0

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	want := mgmtConfigPath(req.WrkHome)
	assert.Output(t, resp.Stdout, v2StdoutTemplate(want))
	if _, statErr := os.Stat(want); !os.IsNotExist(statErr) {
		t.Fatalf("--path must not create config.json at %s (stat err=%v)", want, statErr)
	}
}
```
