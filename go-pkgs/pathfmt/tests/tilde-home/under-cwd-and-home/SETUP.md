# Scenario

**Feature**: path under both process cwd and home uses `~/...`, never cwd-relative

```
# critical distinction from Short
cwd = ~/proj, path = ~/proj/skills/foo
Short     -> "skills/foo"
TildeHome -> "~/proj/skills/foo"
```

## Steps

1. Read process cwd and home (no chdir).
2. Skip when cwd is not a **strict child** of home (not under home, or cwd equals
   home — in the latter case Short also uses `~/...`, so the regression is
   vacuous). Dual membership cannot be forced without forbidden process chdir.
3. Set `req.Path` to an absolute child of cwd (hence also under home) with a
   distinctive segment `marker-cwd-and-home`.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	cwdAbs, _ := mustCwdStrictChildOfHome(t)
	req.Path = filepath.Join(cwdAbs, "skills", "foo-tilde-home", "marker-cwd-and-home")
	return nil
}
```

