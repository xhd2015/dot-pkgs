# Scenario

**Feature**: empty Host is omitted from marshaled Event JSON

```
# omitempty on host
Event{Host:""} -> json.Marshal -> object without "host" key
```

## Steps

1. Clear `req.Event.Host` so the omitempty tag can drop the field.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Event.Host = ""
	return nil
}
```
