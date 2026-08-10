# Scenario

**Feature**: two tools in different dirs → DirsEnv joins with PathListSeparator

```
tool-a in $WorkDir/bin-a, tool-b in $WorkDir/bin-b
  -> DirsEnv contains both dirs separated by os.PathListSeparator
```

## Steps

1. Two ExtraDirs with one executable each; names in that order.
2. LookPath miss.

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
	binA := filepath.Join(req.WorkDir, "bin-a")
	binB := filepath.Join(req.WorkDir, "bin-b")
	writeExecutable(t, filepath.Join(binA, "tool-a"))
	writeExecutable(t, filepath.Join(binB, "tool-b"))
	// Both dirs listed so each name resolves via ExtraDirs scan.
	req.ExtraDirs = []string{binA, binB}
	req.LookPathHits = nil
	return nil
}
```
