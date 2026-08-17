# Scenario

**Feature**: an unclosed fence consumes the rest of the document

```
unclosed Fence -> Document treats later headings as code through EOF
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
	req.Source = "# Target\n\x60\x60\x60\x60\ncode\n# Not a boundary\n\x60\x60\x60\nstill code\n# Also code\n"
	req.Header = "# Target"
	return nil
}
```
