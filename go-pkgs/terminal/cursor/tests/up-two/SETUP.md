# Scenario

**Feature**: Up(2) emits CSI CUU 2

```
Up(2) -> "\x1b[2A"
```

## Steps

1. Set Op `"up"`, N 2.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "up"
	req.N = 2
	return nil
}
```
