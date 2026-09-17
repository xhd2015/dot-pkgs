# Scenario

**Feature**: an empty path is rejected before the runner

```
PathConfig("", cfg) -> error
runner not called
```

## Steps

1. Set `Kind=path`, `GOOS=darwin`, and an empty `Target`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "path"
	req.GOOS = "darwin"
	req.Target = "   "
	return nil
}
```
