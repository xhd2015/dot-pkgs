# Scenario

**Feature**: relative input Abs'd under home becomes `"~/..."`

```
# relative inputs
relative path -> Abs (via process cwd) -> home rules -> "~/..."
```

## Steps

1. Ensure process cwd is under home (skip otherwise; Abs of a relative path
   depends on cwd and chdir is forbidden).
2. Set `req.Path` to a **relative** path with distinctive segment
   `marker-relative-under-home` (not absolute).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_, _ = mustCwdUnderHome(t)
	// Relative on purpose — Abs resolves against process cwd (under home).
	req.Path = filepath.Join("skills", "foo-tilde-home", "marker-relative-under-home")
	if filepath.IsAbs(req.Path) {
		t.Fatalf("setup bug: path must be relative, got %q", req.Path)
	}
	return nil
}
```
