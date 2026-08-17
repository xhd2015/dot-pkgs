# Scenario

**Feature**: empty insertion content creates a heading with an empty body

```
Caller empty content -> Document -> heading line only before boundary
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
	req.Source = "# User\nbody\n"
	req.Header = "# Added"
	req.Content = ""
	return nil
}
```

