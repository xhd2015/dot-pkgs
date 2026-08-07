# Scenario

**Feature**: `ResolveLatestStableURL` — follow redirects to stable zip URL + version

```
LatestURL (fake) -> HTTPClient follows redirects -> final .zip URL + version
```

## Steps

1. Set `Operation=resolve`.
2. Leaves set `HTTPMode` and assert URL/version or error.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "resolve"
	return nil
}
```
