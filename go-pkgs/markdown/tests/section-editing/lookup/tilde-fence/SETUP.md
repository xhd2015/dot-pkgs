# Scenario

**Feature**: a heading-looking line inside a tilde fence is literal content

```
Fence -> Document suppresses fake heading -> Section continues
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
	req.Source = "# Target\n~~~\n# Fake\n~~~\nafter\n# Next\n"
	req.Header = "# Target"
	return nil
}
```

