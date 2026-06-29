---
label: unix
explanation: chmod 000 directory fixture; permission-denied walk skip
---

## Expected

- Exit code 0.
- Stdout contains `visible-repo` discovery line.
- Stderr contains a permission-skip warning mentioning `secret`.

## Exit Code

- `0`.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code = %d, want 0\nstdout:\n%s\nstderr:\n%s", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	roots := rootsFromArgs(req.Args)
	wantPath := absPath(t, filepath.Join(roots[0], "visible-repo"))
	if !strings.Contains(resp.Stdout, wantPath) {
		t.Fatalf("stdout should list visible-repo, got:\n%s", resp.Stdout)
	}
	secretPath := absPath(t, filepath.Join(roots[0], "secret"))
	assert.Output(t, resp.Stderr, `
<contains>
warning: skipping
`+secretPath+`
</contains>`)
}
```