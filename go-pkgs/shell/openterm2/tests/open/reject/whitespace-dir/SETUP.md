# Scenario

**Feature**: whitespace-only dir is rejected before any opener

```
dir=" \t\n " -> OpenConfig -> error
neither opener called
```

## Steps

1. Set `Dir` to a whitespace-only string (spaces, tab, newline).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = " \t\n "
	return nil
}
```
