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
)

func Setup(t *testing.T, req *Request) error {
	// Relative to the leaf directory (doctest cwd).
	req.Path = filepath.Join("testdata", "05-status-fields.snapshot.txt")
	return nil
}
```
