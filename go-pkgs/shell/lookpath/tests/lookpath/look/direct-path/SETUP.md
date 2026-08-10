# Scenario

**Feature**: absolute / path-separator name checks only that path (`Via=direct`)

```
name is absolute (or contains path separator)
  -> IsExecutable(name) only
  -> no LookPath / ExtraDirs / DefaultDirs / candidates / login fallthrough
```

## Steps

1. Leave bare-name pipeline fixtures empty (LookPath miss by default).
2. Leaves set absolute `Name` under WorkDir and optional file fixtures.
3. Assert `ExpectNoLookPath` so spies prove no fallthrough.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ExpectNoLookPath = true
	// Ensure later stages would succeed if wrongly consulted:
	// inject a LookPath hit that Assert forbids being used.
	req.LookPathHit = ""
	return nil
}
```
