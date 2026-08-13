# Scenario

**Feature**: --no-color alone yields Never

```
ModeFromFlags(false, true) -> Never, nil
```

## Steps

1. Set ColorFlag false, NoColorFlag true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ColorFlag = false
	req.NoColorFlag = true
	return nil
}
```
