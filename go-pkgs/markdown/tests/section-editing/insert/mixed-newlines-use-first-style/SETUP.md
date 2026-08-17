# Scenario

**Feature**: inserted bytes use the first observed line-ending style

```
Caller insert -> mixed Document -> first CRLF style for new bytes
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
	req.Source = "lead\r\n# User\nbody\n"
	req.Header = "# Added"
	req.Content = "new"
	return nil
}
```

