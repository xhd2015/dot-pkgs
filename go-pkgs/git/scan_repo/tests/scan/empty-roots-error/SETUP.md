# Scenario

**Feature**: empty roots slice returns validation error

```
len(Roots)==0 -> error: at least one root required
```

## Steps

1. Set `req.Roots` to nil/empty slice.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Roots = nil
	return nil
}
```