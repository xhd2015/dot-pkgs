# Scenario

**Feature**: an absent selector differs from an existing empty section

```
Caller -> Document -> no matching Section
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
	req.Source = "# Existing\n"
	req.Header = "# Missing"
	return nil
}
```

