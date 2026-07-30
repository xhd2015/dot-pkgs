# Scenario

**Feature**: untracked file dirties go-pkgs porcelain clean APIs

```
# write untracked.txt only (no git add)
IsClean=error, IsCleanWrk=false
```

## Preconditions

- Clean repo from clean/ grouping bootstrap.

## Steps

1. Write untracked file `untracked.txt` (do not `git add`).

## Context

- Scenario 3 from P2: porcelain dirty with untracked via go-pkgs names.
- `IsClean` treats any non-empty porcelain as dirty (error).
- `IsCleanWrk` maps `??` → added → dirty (false).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	t.Helper()
	writeFile(t, req.Dir, "untracked.txt", "only untracked\n")
	return nil
}
```
