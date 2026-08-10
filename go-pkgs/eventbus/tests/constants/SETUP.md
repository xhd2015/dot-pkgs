# Scenario

**Feature**: locked wire constants for port, event types, and sources

```
# package constants are the v1 wire vocabulary
caller -> eventbus.DefaultPublishPort / Type* / Source* -> fixed strings
```

## Steps

1. Set `req.Op` to `"constants"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "constants"
	return nil
}
```
