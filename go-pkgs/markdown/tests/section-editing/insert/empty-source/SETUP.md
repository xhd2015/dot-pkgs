# Scenario

**Feature**: insertion into an empty document starts with the new heading

```
Caller insert -> empty Document -> new Section
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
	req.Source = ""
	req.Header = "# Added"
	req.Content = "body"
	return nil
}
```

