# Scenario

**Feature**: Up(0) is empty

```
Up(0) -> ""
```

## Steps

1. Set Op `"up"`, N 0.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "up"
	req.N = 0
	return nil
}
```
