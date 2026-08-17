# Scenario

**Feature**: an invalid seven-level selector is rejected without mutation

```
Caller invalid selector -> Document sentinel error -> original bytes
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
	req.Header = "####### Too Deep"
	return nil
}
```

