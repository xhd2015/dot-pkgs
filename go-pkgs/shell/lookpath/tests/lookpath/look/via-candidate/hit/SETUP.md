# Scenario

**Feature**: ExtraCandidates absolute path wins

```
ExtraCandidates=[$WorkDir/cand/mytool] executable
  -> Path=.../cand/mytool, Via=candidate
```

## Steps

1. Write executable at `$WorkDir/cand/mytool`.
2. Set `ExtraCandidates` to that absolute path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	cand := filepath.Join(req.WorkDir, "cand", "mytool")
	writeExecutable(t, cand)
	req.ExtraCandidates = []string{cand}
	return nil
}
```
