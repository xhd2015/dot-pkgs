# Scenario

**Feature**: Download 200 with empty body → error

```
GET /file.zip -> 200 empty -> error
```

## Steps

1. Set `HTTPMode=download-empty`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.HTTPMode = "download-empty"
	return nil
}
```
