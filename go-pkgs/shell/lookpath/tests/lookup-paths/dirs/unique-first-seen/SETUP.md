# Scenario

**Feature**: two tools in the same bin dir collapse to one Dir entry

```
Names=["a","b"] both under $WorkDir/samebin via ExtraDirs
  -> Dirs() == [$WorkDir/samebin] (one entry, first-seen)
```

## Steps

1. Create two executables in one ExtraDir.
2. LookPath miss so ExtraDirs wins.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"tool-a", "tool-b"}
	same := filepath.Join(req.WorkDir, "samebin")
	writeExecutable(t, filepath.Join(same, "tool-a"))
	writeExecutable(t, filepath.Join(same, "tool-b"))
	req.ExtraDirs = []string{same}
	req.LookPathHits = nil
	return nil
}
```
