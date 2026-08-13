# Scenario

**Feature**: --color alone yields Always

```
ModeFromFlags(true, false) -> Always, nil
```

## Steps

1. Set ColorFlag true, NoColorFlag false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ColorFlag = true
	req.NoColorFlag = false
	return nil
}
```
