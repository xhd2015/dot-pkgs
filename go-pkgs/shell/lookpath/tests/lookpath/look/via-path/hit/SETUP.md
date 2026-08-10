# Scenario

**Feature**: injected LookPath returns an absolute path

```
opts.LookPath("mytool") -> /injected/bin/mytool
Look -> Path=/injected/bin/mytool, Via=path
```

## Steps

1. Set `LookPathHit` to a synthetic absolute path under WorkDir.
2. Do not create ExtraDirs / candidates (unused on PATH hit).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LookPathHit = filepath.Join(req.WorkDir, "injected", "bin", "mytool")
	// File need not exist for this stage if product trusts LookPath result;
	// still create exec so IsExecutable checks (if product re-validates) pass.
	writeExecutable(t, req.LookPathHit)
	return nil
}
```
