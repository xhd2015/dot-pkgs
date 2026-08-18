# Scenario

**Feature**: production TTY status snapshot is UTF-8 text, not binary

```
testdata/05-status-fields.snapshot.txt -> DetectFileType -> isBinary=false
```

## Steps

1. Point `req.Path` at the internalized production fixture
   `testdata/05-status-fields.snapshot.txt` (byte-identical to the agent TTY snapshot
   that was auto-unstaged as binary).
2. Run `DetectFileType` via root `Run`.

```go
import (
	"path/filepath"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	// Relative to the leaf directory (doctest cwd).
	caseDir := d.DOCTEST_CASE
	if caseDir == "" || !filepath.IsAbs(caseDir) {
		caseDir = filepath.Join(d.DOCTEST_ROOT, caseDir)
	}
	req.Path = filepath.Join(caseDir, "testdata", "05-status-fields.snapshot.txt")
	return nil
}
```
