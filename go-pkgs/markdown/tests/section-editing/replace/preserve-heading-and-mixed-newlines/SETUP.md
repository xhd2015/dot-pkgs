# Scenario

**Feature**: replacement preserves heading spelling and surrounding mixed line endings

```
Caller replacement -> selected body only -> untouched mixed bytes
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
	req.Source = "lead\n   # Target ###\r\nold\r\n# Next\nkeep\n"
	req.Header = "# Target"
	req.Content = "new"
	return nil
}
```

