# Scenario

**Feature**: darwin opens a target with the default handler

```
Args(target, "darwin") -> ["open", target]
```

## Steps

1. Set `Kind=path`, `GOOS=darwin`.
2. `Target` is the existing fixture directory from root Setup.

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
	return nil
}
```
