# Scenario

**Feature**: injected LookPath resolves bare name; From empty

```
opts.LookPath("mytool") -> /…/injected/bin/mytool
LookupPaths -> Item{Name=mytool, Path=…, Missing=false, From=""}
```

## Steps

1. Single name `mytool`.
2. Set LookPathHits for `mytool` to a synthetic absolute path; create executable.

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
	hit := filepath.Join(req.WorkDir, "injected", "bin", "mytool")
	writeExecutable(t, hit)
	req.LookPathHits = map[string]string{"mytool": hit}
	return nil
}
```
