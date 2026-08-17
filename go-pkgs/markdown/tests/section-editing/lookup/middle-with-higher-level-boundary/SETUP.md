# Scenario

**Feature**: a level-three section in the middle stops at a higher-level heading

```
Caller -> Document -> middle body -> higher-level boundary
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
	req.Source = "preamble\n# Parent\n## Area\n### Target\nvalue\n# Final\nend\n"
	req.Header = "### Target"
	return nil
}
```

