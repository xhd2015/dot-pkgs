# Scenario

**Feature**: removing a middle section joins untouched preceding and following bytes

```
Caller remove -> middle Section disappears -> neighbors meet
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
	req.Source = "lead\n# A\nx\n# B\nb\n## Child\nc\n# C\nz\n"
	req.Header = "# B"
	return nil
}
```

