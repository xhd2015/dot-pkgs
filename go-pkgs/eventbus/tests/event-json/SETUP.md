# Scenario

**Feature**: Event JSON envelope marshal and unmarshal

```
# standard encoding/json on eventbus.Event
Event struct -> json.Marshal -> bytes -> json.Unmarshal -> Event
```

## Steps

1. Set `req.Op` to `"event-json"`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "event-json"
	return nil
}
```
