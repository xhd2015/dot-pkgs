# Scenario

**Feature**: custom `Options.OpenFiles` paths are returned as-is for the target pid

```
OpenFilesInject ["/tmp/a","/tmp/b"] for pid 4242 -> OpenFiles -> same paths
```

## Steps

1. Set `req.OpenFilesPID=4242` and inject two absolute paths.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.OpenFilesPID = 4242
	req.OpenFilesInject = []string{"/tmp/a", "/tmp/b"}
	return nil
}
```
