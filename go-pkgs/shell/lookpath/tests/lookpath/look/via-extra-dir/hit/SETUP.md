# Scenario

**Feature**: first ExtraDirs entry containing the binary wins

```
ExtraDirs=[$WorkDir/extra] + mytool -> Path=.../extra/mytool, Via=extra_dir
```

## Steps

1. Create `$WorkDir/extra/mytool` as executable.
2. Set `ExtraDirs` to that directory.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	extra := filepath.Join(req.WorkDir, "extra")
	writeExecutable(t, filepath.Join(extra, "mytool"))
	req.ExtraDirs = []string{extra}
	return nil
}
```
