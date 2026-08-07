# Scenario

**Feature**: Download HTTP 404 → error

```
GET /file.zip -> 404 -> error
```

## Steps

1. Set `HTTPMode=download-404`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.HTTPMode = "download-404"
	return nil
}
```
