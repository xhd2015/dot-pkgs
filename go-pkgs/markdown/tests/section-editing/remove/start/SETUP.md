# Scenario

**Feature**: removing the first section leaves the adjacent section byte-exact

```
Caller remove -> first Section disappears -> next Section retained
```

## Steps

1. Provide the exact source, selector, and content needed by this case.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Source = "# A\nx\n# B\r\ny\r\n"
	req.Header = "# A"
	return nil
}
```

