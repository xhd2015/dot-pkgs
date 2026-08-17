# Scenario

**Feature**: insertion rejects a malformed selector without changing an empty document

```
Caller invalid selector -> Document sentinel error -> empty bytes retained
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
	req.Source = ""
	req.Header = "#NoSpace"
	req.Content = "new"
	return nil
}
```

