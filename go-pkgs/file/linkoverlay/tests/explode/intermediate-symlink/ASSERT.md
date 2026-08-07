## Expected

- Apply succeeds.
- `target/.config` is a real directory (not a symlink) after explode.
- `target/.config/other/x` still readable as `O` (sibling re-linked from base).
- `target/.config/tool/c` is regular file content `C`.
- Base `.config/other/x` still `O` on disk.

## Errors

- No error.

```go
import (
	"os"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	_ = d
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	cfg := targetPath(req, resp, ".config")
	st := mustLstat(t, cfg)
	if st.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("target/.config should be a real directory after explode, still symlink")
	}
	if !st.IsDir() {
		t.Fatalf("target/.config: want directory, mode %v", st.Mode())
	}
	assertRegularContent(t, targetPath(req, resp, ".config/other/x"), "O")
	tool := targetPath(req, resp, ".config/tool/c")
	stTool := mustLstat(t, tool)
	if stTool.Mode()&os.ModeSymlink != 0 {
		t.Fatalf("target/.config/tool/c should be regular file")
	}
	assertRegularContent(t, tool, "C")
	assertFileContentUnchanged(t, basePath(req, "base0", ".config/other/x"), "O")
}
```
