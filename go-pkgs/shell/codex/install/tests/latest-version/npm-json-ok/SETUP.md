# Scenario

**Feature**: fake npm latest JSON returns version string

```
HTTP 200 {"version":"0.147.0"} -> LatestVersion -> "0.147.0"
```

## Steps

1. Set `HTTPMode=npm-ok` and fixture version `0.147.0`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.HTTPMode = "npm-ok"
	req.NPMVersion = "0.147.0"
	return nil
}
```
