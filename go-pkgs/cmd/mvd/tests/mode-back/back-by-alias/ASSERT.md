## Expected
- The project referenced by the alias `opencode` is moved back to its original location at `projects/opencode-latest`.
- The output contains "moved back", confirming the back operation succeeded via alias resolution.
- The moved location `scratch/opencode-latest` no longer exists.
- The history chain is reduced to a single entry at `projects/opencode-latest` (back at origin).

## Exit Code
- 0 (success)

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
    if resp == nil {
        t.Fatalf("expected response, got error: %v", err)
    }
    if resp.ExitCode != 0 {
        t.Fatalf("exit code: %d, output:\n%s", resp.ExitCode, resp.Output)
    }
    assertContains(t, resp.Output, "moved back")
    
    src := filepath.Join(req.WorkRoot, "projects", "opencode-latest")
    dstPath := filepath.Join(req.WorkRoot, "scratch", "opencode-latest")
    assertFileExists(t, src)
    assertFileNotExists(t, dstPath)
    
    assertHistoryChain(t, req.ConfigHome, src, src)
}
```
