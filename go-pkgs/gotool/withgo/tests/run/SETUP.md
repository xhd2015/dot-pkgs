# Scenario

**Feature**: Run resolves an existing dest then Execs fake `$dest/bin/go`

```
# compose ResolveGoroot + Exec
go1.19 + existing $InstallDir/go1.19.13/bin/go -> Run -> child GOROOT=$dest
```

## Steps

1. Set `req.Op` to `run`, `GoVersion` to `go1.19`.
2. Inject InstallDir + recording hook. Create dest dir and fake `bin/go`.
3. Set args to `["go"]` and ExtraEnv `WITHGO_EXTRA=from-test`.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "run"
	req.GoVersion = "go1.19"
	req.InstallDir = t.TempDir()
	req.RecordInstall = true
	req.Download = false
	dest := filepath.Join(req.InstallDir, "go1.19.13")
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	writeFakeGo(t, dest)
	req.Args = []string{"go"}
	req.ExtraEnv = []string{"WITHGO_EXTRA=from-test"}
	return nil
}
```
