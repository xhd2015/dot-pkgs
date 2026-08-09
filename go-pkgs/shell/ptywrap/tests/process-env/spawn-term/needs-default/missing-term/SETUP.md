# Scenario

**Feature**: no TERM in final merge → TERM=xterm-256color (S6 missing)

```
# missing TERM
Base has PATH only (no TERM); empty set/unset
  -> MergeProcessEnv + EnsureSpawnTERM
  -> TERM=xterm-256color; PATH preserved
```

## Steps

1. Base: `PATH=/bin`, `HOME=/tmp/home` (no TERM).
2. Empty Set and Unset.
3. ApplySpawnTERM true (from ancestors).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Base = []string{
		"PATH=/bin",
		"HOME=/tmp/home",
	}
	req.Set = nil
	req.Unset = nil
	return nil
}
```
