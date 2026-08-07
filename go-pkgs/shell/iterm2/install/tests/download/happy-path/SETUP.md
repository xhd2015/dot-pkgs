# Scenario

**Feature**: Download happy path writes non-empty zip to DestPath

```
GET /file.zip (fixture bytes) -> dest exists, size > 0
```

## Steps

1. Set `HTTPMode=download-ok`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.HTTPMode = "download-ok"
	return nil
}
```
