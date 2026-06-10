## Expected
- When the destination directory exists, mvd moves the source into it (not renames to it).
- The file is at `existing-dir/mysrc/f.txt`.
- History chain records the move into the subdirectory.

## Exit Code
- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	src := filepath.Join(req.WorkRoot, "mysrc")
	dst := filepath.Join(req.WorkRoot, "existing-dir")
	moved := filepath.Join(dst, "mysrc")
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d: %s", resp.ExitCode, resp.Output)
	}
	assertFileExists(t, filepath.Join(moved, "f.txt"))
	assertFileNotExists(t, src)
	assertHistoryChain(t, req.ConfigHome, src, src, moved)
}
```
