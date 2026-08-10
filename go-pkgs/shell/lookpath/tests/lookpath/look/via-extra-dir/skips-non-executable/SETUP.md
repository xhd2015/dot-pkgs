# Scenario

**Feature**: non-executable match in first ExtraDir is skipped; second dir wins

```
ExtraDirs=[d1,d2]
  d1/mytool (0644) skipped
  d2/mytool (0755) hit -> Via=extra_dir, Path=d2/mytool
```

## Steps

1. Create non-executable in first ExtraDir and executable in second.
2. Set `ExtraDirs` to both directories in order.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	d1 := filepath.Join(req.WorkDir, "extra1")
	d2 := filepath.Join(req.WorkDir, "extra2")
	writeNonExecutable(t, filepath.Join(d1, "mytool"))
	writeExecutable(t, filepath.Join(d2, "mytool"))
	req.ExtraDirs = []string{d1, d2}
	return nil
}
```
