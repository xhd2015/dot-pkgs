# Scenario

**Feature**: an unsupported platform is rejected instead of guessing

```
Args(target, "plan9") -> error
```

## Steps

1. Set `Kind=path` and an unsupported `GOOS`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "path"
	req.GOOS = "plan9"
	return nil
}
```
