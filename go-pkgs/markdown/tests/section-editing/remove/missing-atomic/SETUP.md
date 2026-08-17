# Scenario

**Feature**: removing a missing section returns a sentinel and is atomic

```
Caller missing selector -> Document error -> original bytes
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
	req.Source = "# Existing\nbody\n"
	req.Header = "# Missing"
	return nil
}
```

