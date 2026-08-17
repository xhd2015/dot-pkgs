# Scenario

**Feature**: insertion preserves frontmatter and leading prose before the first heading

```
Caller insert -> Document keeps preamble -> new Section before first Section
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
	req.Source = "---\ntitle: x\n---\nlead\n# User\nbody\n"
	req.Header = "# Managed"
	req.Content = "rules"
	return nil
}
```

