# Scenario

**Feature**: remove one complete selected section

```
Caller -> Document remove -> Section disappears
Document -> Caller adjacent sections preserved
```

## Steps

1. Select the remove operation.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "remove"
	return nil
}
```
