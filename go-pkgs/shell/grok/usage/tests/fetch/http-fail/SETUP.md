# Scenario

**Feature**: Non-auth billing GET failure

```
GetJSON=HTTP 500 -> error
```

## Steps

1. Set `FetchMode=http-fail`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "http-fail"
	return nil
}
```
