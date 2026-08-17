# Scenario

**Feature**: a heading at EOF is present with an empty body

```
Caller -> Document -> present empty Section at EOF
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
	req.Source = "# First\nx\n# Empty"
	req.Header = "# Empty"
	return nil
}
```

