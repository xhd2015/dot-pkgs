# Scenario

**Feature**: inserting an already existing selector returns a sentinel and is atomic

```
Caller existing selector -> Document error -> no duplicate created
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
	req.Source = "# Added\nbody\n# User\n"
	req.Header = "# Added"
	req.Content = "new"
	return nil
}
```

