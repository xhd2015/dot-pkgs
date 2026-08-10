# Scenario

**Feature**: unset TERM then spawn policy fills default

```
# unset TERM interaction
Base has TERM=xterm and PATH=/bin
Unset = TERM
  -> MergeProcessEnv removes TERM
  -> EnsureSpawnTERM sets TERM=xterm-256color
```

## Steps

1. Base: `TERM=xterm`, `PATH=/bin`.
2. Unset: `TERM`.
3. Set empty.
4. ApplySpawnTERM true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Base = []string{
		"TERM=xterm",
		"PATH=/bin",
	}
	req.Unset = []string{"TERM"}
	req.Set = nil
	return nil
}
```
