# Scenario

**Feature**: Publish posts Event JSON to POST /publish without Authorization when token empty

```
# canonical success path
Publisher(baseURL, token="") -> POST /publish Content-Type application/json body=Event
Authorization header absent
```

## Steps

1. Leave `req.Token` empty (default).
2. Keep mock BaseURL from parent success Setup.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Token = ""
	return nil
}
```
