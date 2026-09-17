# Scenario

**Feature**: a runner failure surfaces instead of a Result

```
PathConfig(path, cfg with Run error) -> error, no Result
```

## Steps

1. Set `Kind=path`, `GOOS=darwin`, and `RunnerErr`.

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
	req.RunnerErr = "open: exit status 1"
	return nil
}
```
