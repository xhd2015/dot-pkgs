# Scenario

**Feature**: neither flag yields Auto

```
ModeFromFlags(false, false) -> Auto, nil
```

## Steps

1. Set both flags false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ColorFlag = false
	req.NoColorFlag = false
	return nil
}
```
