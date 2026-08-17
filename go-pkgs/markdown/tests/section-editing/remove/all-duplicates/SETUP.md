# Scenario

**Feature**: remove every duplicate section while retaining unrelated sections

```
Caller duplicate selector -> Document remove all -> matching Sections disappear
Document -> Caller unrelated Sections remain byte-exact
```

## Steps

1. Select remove-all with duplicate managed-looking sections around a user section.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "remove_all"
	req.Source = "# Legacy\none\n## Child\nold\n# User\nkeep\n# Legacy\ntwo\n"
	req.Header = "# Legacy"
	return nil
}
```
