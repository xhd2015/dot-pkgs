# Scenario

**Feature**: SpaceAllow space-first filter (FixedSpace; no live CGS)

```
two windows with FixedSpace 0 and 2
  -> CaptureWith(SpaceAllow) keeps/skips deep-capture by allowlist
```

## Preconditions

- Fixture hierarchy; FixedSpace set on windows (ResolveSpace unused).
- Both windows have idle sessions so enrich is deterministic when kept.

## Steps

1. Parent zeros via root Setup; leaves set Windows + SpaceAllow + IdleTTYs.
2. Run uses CaptureWith(SpaceAllow) and records SpaceSkipped.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.ITermRunning = true
	return nil
}
```
