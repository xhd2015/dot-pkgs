# Scenario

**Feature**: already-full patch versions are identity with a `go` prefix

```
# go1.19.13 and naked 1.19.13 both stay go1.19.13
go1.19.13 -> PinPatch -> go1.19.13
1.19.13 -> PinPatch -> go1.19.13
```

## Steps

1. Set `req.PinInputs` to `go1.19.13` and `1.19.13`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.PinInputs = []string{"go1.19.13", "1.19.13"}
	return nil
}
```
