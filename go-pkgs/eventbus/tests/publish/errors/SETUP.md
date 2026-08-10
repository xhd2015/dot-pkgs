# Scenario

**Feature**: Publisher.Publish surfaces HTTP and transport failures as errors

```
# best-effort callers still observe errors from Publish
non-2xx response -> error
transport failure -> error
```

## Steps

1. Set `req.Op` remains `"publish"` from parent.
2. Leaves configure mock failure mode (status code or closed server).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "publish"
	req.Capture = &HTTPCapture{}
	return nil
}
```
