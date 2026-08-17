# Scenario

**Feature**: unknown major.minor is left unchanged with a `go` prefix

```
# go1.99 is not in the kool pin map
go1.99 -> PinPatch -> go1.99
```

## Steps

1. Set `req.GoVersion` to `go1.99`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.GoVersion = "go1.99"
	return nil
}
```
