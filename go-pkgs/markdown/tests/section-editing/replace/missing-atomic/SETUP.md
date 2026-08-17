# Scenario

**Feature**: replacing a missing section returns a sentinel and preserves all bytes

```
Caller missing selector -> Document error -> no mutation
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
	req.Source = "preamble\n# Existing\nbody\n"
	req.Header = "# Missing"
	req.Content = "new"
	return nil
}
```

