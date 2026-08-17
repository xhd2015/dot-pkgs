# Scenario

**Feature**: removing the final section preserves the preceding document exactly

```
Caller remove -> EOF Section disappears -> prefix retained
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
	req.Source = "# A\r\nx\r\n## Last\r\ny"
	req.Header = "## Last"
	return nil
}
```

