# Scenario

**Feature**: a section at document start includes its lower-level child and stops at its next peer

```
Caller -> Document -> body with child heading -> same-level boundary
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
	req.Source = "# Alpha\nbody\n## Child\nchild body\n# Beta\nkeep\n"
	req.Header = "# Alpha"
	return nil
}
```

