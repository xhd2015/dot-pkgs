## Expected

- `err == nil`.
- `resp.BackupPath` non-empty and basename starts with `iTerm.app.bak-`.
- Backup MacOS binary still contains `OLD-INSTALL`.
- New target MacOS binary does **not** contain `OLD-INSTALL`.

## Errors

- None.

```go
import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	t.Helper()
	assertNoError(t, err)
	if resp.BackupPath == "" {
		t.Fatal("expected BackupPath for replaced install")
	}
	if !strings.HasPrefix(filepath.Base(resp.BackupPath), "iTerm.app.bak-") {
		t.Fatalf("backup name %q want prefix iTerm.app.bak-", filepath.Base(resp.BackupPath))
	}
	assertDirExists(t, resp.BackupPath)
	old := readFileString(t, filepath.Join(resp.BackupPath, "Contents", "MacOS", "iTerm2"))
	if !strings.Contains(old, req.ExistingMarker) {
		t.Fatalf("backup missing marker %q: %q", req.ExistingMarker, old)
	}
	neu := readFileString(t, filepath.Join(resp.AppPath, "Contents", "MacOS", "iTerm2"))
	if strings.Contains(neu, req.ExistingMarker) {
		t.Fatalf("new install still has old marker: %q", neu)
	}
}
```
