# Scenario

**Feature**: HTTP error from latest endpoint → error

```
GET /stable/latest -> 500 -> error
```

## Steps

1. Set `HTTPMode=http-error`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.HTTPMode = "http-error"
	return nil
}
```
