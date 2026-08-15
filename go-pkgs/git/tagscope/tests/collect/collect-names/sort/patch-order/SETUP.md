# Scenario

**Feature**: patch components compare numerically in version sort

```
[v0.0.1, v0.0.10, v0.0.2] -> Tags v0.0.10, v0.0.2, v0.0.1
```

## Steps

1. Set `req.Names` with non-lexicographic patch ordering.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.1", "v0.0.10", "v0.0.2"}
	return nil
}
```