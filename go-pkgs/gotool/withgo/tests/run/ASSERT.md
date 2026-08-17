## Expected

- `Run` succeeds (resolve existing dest, then exec fake go).
- Install hook is unused.
- Child GOROOT is abs dest `$InstallDir/go1.19.13`; PATH0 is `$dest/bin`.

## Expected Output

```
---
version: 3
__GOROOT__: type=string, example=/tmp/installed/go1.19.13
__BIN__: type=string, example=/tmp/installed/go1.19.13/bin
---
GOROOT=__GOROOT__
PATH0=__BIN__
WITHGO_EXTRA=from-test
```

## Errors

- `err` and `resp.Err` are nil.

```go
import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected run error: %v", err)
	}
	if resp.Err != nil {
		t.Fatalf("Run(%q, %v) failed: %v", req.GoVersion, req.Args, resp.Err)
	}
	if resp.HookCalled {
		t.Fatalf("Install hook called (version=%q dir=%q); dest already existed", resp.HookVersion, resp.HookDir)
	}
	dest := absPath(t, filepath.Join(req.InstallDir, "go1.19.13"))
	bin := filepath.Join(dest, "bin")
	assert.Output(t, resp.Stdout, `---
version: 3
---
GOROOT=`+regexp.QuoteMeta(dest)+`
PATH0=`+regexp.QuoteMeta(bin)+`
WITHGO_EXTRA=from-test
`)
}
```
