# Scenario

**Feature**: empty input is returned unchanged by `Expand`

```
# Expand passthrough
empty -> unchanged
```

## Steps

1. Set `req.Path` to `""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Path = ""
	return nil
}```
