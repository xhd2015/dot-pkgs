# Scenario

**Feature**: replacing with the existing body is byte-for-byte idempotent

```
Caller existing body -> Document replace -> identical repeated String
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
	req.Source = "# Target\r\nsame\r\n## Child\nchild\r\n# Next\r\n"
	req.Header = "# Target"
	req.Content = "same\r\n## Child\nchild\r\n"
	return nil
}
```

