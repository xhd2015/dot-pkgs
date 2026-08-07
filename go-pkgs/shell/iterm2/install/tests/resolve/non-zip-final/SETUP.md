# Scenario

**Feature**: final redirect target is not a zip → error

```
GET /stable/latest -> 302 /downloads/not-a-zip.html -> error
```

## Steps

1. Set `HTTPMode=non-zip`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.HTTPMode = "non-zip"
	return nil
}
```
