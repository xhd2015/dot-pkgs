# Scenario

**Feature**: dry-run lists fetch + rebase without mutating

```
DryRun=true -> stdout contains fetch and rebase origin/
```

```go
import (
	"bytes"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.DryRun = true
	req.Stdout = &bytes.Buffer{}
	return nil
}
```
