# Scenario

**Feature**: `NeedsUpdate` — pure semver local vs latest compare

```
local, latest -> NeedsUpdate -> true only if both parseable and local < latest
```

## Steps

1. Set `Operation=needs-update`.
2. Leaves set `LocalVer` / `LatestVer` pairs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "needs-update"
	return nil
}
```
