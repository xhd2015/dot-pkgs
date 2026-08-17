# Scenario

**Feature**: insertion appends after heading-free source and supplies one separator newline

```
Caller insert -> heading-free Document -> append Section
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
	req.Source = "preamble"
	req.Header = "# Added"
	req.Content = "body"
	return nil
}
```

