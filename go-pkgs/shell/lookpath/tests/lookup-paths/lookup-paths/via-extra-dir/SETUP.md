# Scenario

**Feature**: ExtraDirs resolves bare name after PATH miss; From empty

```
LookPath miss + ExtraDirs=[$WorkDir/extra]/mytool
  -> Item found, Path=…/extra/mytool, From=""
```

## Steps

1. Create `$WorkDir/extra/mytool` executable; set ExtraDirs.
2. Leave LookPathHits empty (PATH miss).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"mytool"}
	extra := filepath.Join(req.WorkDir, "extra")
	writeExecutable(t, filepath.Join(extra, "mytool"))
	req.ExtraDirs = []string{extra}
	req.LookPathHits = nil
	return nil
}
```
