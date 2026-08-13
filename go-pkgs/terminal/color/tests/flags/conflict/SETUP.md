# Scenario

**Feature**: both flags together are an error

```
ModeFromFlags(true, true) -> error "--color and --no-color cannot be specified together"
```

## Steps

1. Set both ColorFlag and NoColorFlag true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ColorFlag = true
	req.NoColorFlag = true
	return nil
}
```
