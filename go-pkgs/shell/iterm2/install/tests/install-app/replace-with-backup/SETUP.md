# Scenario

**Feature**: replace existing target creates `.bak-<unix_ts>` backup

```
existing target (marker=OLD) + new extracted
  -> target has new content; sibling iTerm.app.bak-<ts> retains OLD
```

## Steps

1. Seed existing target with marker `OLD-INSTALL`.
2. Install fresh extracted app over it.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.SeedExistingTarget = true
	req.ExistingMarker = "OLD-INSTALL"
	req.UseDefaultTarget = false
	return nil
}
```
