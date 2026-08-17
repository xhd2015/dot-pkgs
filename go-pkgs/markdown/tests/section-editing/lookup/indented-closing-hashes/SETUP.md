# Scenario

**Feature**: indented ATX headings with optional closing hashes match normalized selectors

```
Caller normalized selector -> indented Document heading -> exact body
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
	req.Source = "   ### Detail ###   \r\nvalue\r\n  ## Boundary ##\r\n"
	req.Header = "### Detail"
	return nil
}
```

