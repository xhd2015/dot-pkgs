# Scenario

**Feature**: replacing a duplicate section is ambiguous and atomic

```
Caller duplicate selector -> Document error -> neither body changes
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
	req.Content = "new"
	return nil
}
```

