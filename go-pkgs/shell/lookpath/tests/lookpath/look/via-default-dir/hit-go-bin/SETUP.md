# Scenario

**Feature**: binary under `$HOME/go/bin` resolves via default dirs

```
Home=$WorkDir/home
  $Home/go/bin/mytool executable
  -> Path=.../go/bin/mytool, Via=default_dir
```

## Steps

1. Set `Home` to `$WorkDir/home`.
2. Write executable at `$Home/go/bin/mytool`.
3. Leave ExtraCandidates empty; RunLogin always fails (default).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	home := filepath.Join(req.WorkDir, "home")
	req.Home = home
	writeExecutable(t, filepath.Join(home, "go", "bin", "mytool"))
	return nil
}
```
