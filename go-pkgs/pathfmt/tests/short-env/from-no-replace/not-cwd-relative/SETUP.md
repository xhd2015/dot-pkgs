# Scenario

**Feature**: `ShortEnvFrom` never produces cwd-relative forms (unlike `Short`)

```
# when cwd and path both under home, empty env
path under cwd+home -> ~/...  (not "child/..." or ".")
```

## Steps

1. Skip if process cwd is not under home (cannot force dual membership without chdir).
2. Empty env; path is a child of cwd (still under home).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	cwdAbs, _ := mustCwdUnderHome(t)
	req.Env = []string{}
	req.Path = filepath.Join(cwdAbs, "doctest-short-env-not-cwd-rel-marker")
	return nil
}
```
