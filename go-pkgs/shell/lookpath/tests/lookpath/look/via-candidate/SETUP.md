# Scenario

**Feature**: bare name resolved via `opts.ExtraCandidates` → `Via=candidate`

```
LookPath + ExtraDirs + DefaultDirs miss
  -> ExtraCandidates absolute paths -> first executable -> Via=candidate
```

## Steps

1. Force LookPath miss; leave ExtraDirs empty; Home empty so DefaultDirs has no home bins.
2. Leaves set ExtraCandidates.

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
	req.Home = ""
	return nil
}
```
