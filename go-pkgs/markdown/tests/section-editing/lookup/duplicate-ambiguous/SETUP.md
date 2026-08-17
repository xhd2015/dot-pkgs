# Scenario

**Feature**: duplicate matching headings make lookup ambiguous

```
Caller selector -> two Sections -> ambiguous sentinel
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
	req.Source = "# Same\none\n# Same\ntwo\n"
	req.Header = "# Same"
	return nil
}
```

