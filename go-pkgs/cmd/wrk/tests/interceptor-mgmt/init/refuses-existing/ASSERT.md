## Expected

- Non-zero exit code.
- Config file still has argv starting with `custom-tool` (unchanged).
- Prefer empty stdout.

## Errors

- Interceptor already present; hint `--force` optional in stderr.

## Exit Code

- non-zero

```go
import (
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit when interceptor exists without --force")
	}
	ic := readMgmtInterceptor(t, req.WrkHome)
	if ic == nil {
		t.Fatal("interceptor block missing after refused init")
	}
	if len(ic.Argv) < 1 || ic.Argv[0] != "custom-tool" {
		t.Fatalf("config should be unchanged; argv=%v", ic.Argv)
	}
	if !ic.Enabled {
		t.Fatal("enabled should remain true")
	}
	// Soft message about existing / force
	assert.Output(t, resp.Stderr, `<contains>
exist
</contains>`)
}
```
