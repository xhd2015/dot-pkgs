# Scenario

**Feature**: bare name resolved under `opts.ExtraDirs` → `Via=extra_dir`

```
LookPath miss -> scan ExtraDirs+name -> first executable -> Via=extra_dir
```

## Steps

1. Force LookPath miss (default).
2. Leaves place executables under ExtraDirs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Name = "mytool"
	req.LookPathHit = ""
	return nil
}
```
