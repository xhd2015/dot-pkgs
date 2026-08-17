# Scenario

**Feature**: replacing an empty EOF body supplies the document line ending

```
Caller empty replacement -> Document heading at EOF -> terminated empty Section
```

## Steps

1. Use a matching heading at EOF with no line ending.
2. Replace its existing empty body with empty content.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Source = "# Target"
	req.Header = "# Target"
	req.Content = ""
	return nil
}
```
