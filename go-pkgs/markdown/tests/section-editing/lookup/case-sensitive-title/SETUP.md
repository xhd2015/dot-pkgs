# Scenario

**Feature**: heading title matching is case-sensitive

```
Caller differently-cased selector -> Document -> no match
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
	req.Source = "# Title\nbody\n"
	req.Header = "# title"
	return nil
}
```

