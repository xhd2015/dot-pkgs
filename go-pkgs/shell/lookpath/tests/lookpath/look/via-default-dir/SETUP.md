# Scenario

**Feature**: bare name resolved under `DefaultDirs(home)` → `Via=default_dir`

```
LookPath miss, ExtraDirs empty
  -> DefaultDirs(Home)+name -> first executable -> Via=default_dir
```

## Steps

1. Force LookPath miss; leave ExtraDirs empty.
2. Leaves set Home and place binary under a default dir.

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
	req.ExtraDirs = nil
	return nil
}
```
