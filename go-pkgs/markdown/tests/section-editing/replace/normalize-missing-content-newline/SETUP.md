# Scenario

**Feature**: replacement content without a final newline gains the document line ending

```
Caller unterminated content -> Document newline policy -> normalized body
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
	req.Source = "# Target\nold\n# Next\n"
	req.Header = "# Target"
	req.Content = "line1\nline2"
	return nil
}
```

