# Scenario

**Feature**: home set includes home-relative bins and system bins

```
DefaultDirs("/tmp/home-like") includes:
  $home/.local/bin, $home/go/bin, /opt/homebrew/bin, /usr/local/bin
```

## Steps

1. Set `DefaultDirsHome` to `$WorkDir/home` (path need not exist).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DefaultDirsHome = filepath.Join(req.WorkDir, "home")
	return nil
}
```
