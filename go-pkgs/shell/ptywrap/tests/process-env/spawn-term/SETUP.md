# Scenario

**Feature**: spawn path applies `EnsureSpawnTERM` after pure merge

```
# spawn-term surface
req.ApplySpawnTERM = true
  -> Run: EnsureSpawnTERM(MergeProcessEnv(base, set, unset))
  -> TERM defaulted only when missing, empty, or dumb
```

## Preconditions

- This branch exercises spawn TERM policy on top of merge.
- Base/Set/Unset remain injectable (no process environ).

## Steps

1. Set `req.ApplySpawnTERM` to true.

## Context

- Default TERM value is exactly `xterm-256color`.
- Good TERM values are never replaced.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ApplySpawnTERM = true
	return nil
}
```
