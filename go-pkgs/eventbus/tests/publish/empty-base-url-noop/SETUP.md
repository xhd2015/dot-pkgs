# Scenario

**Feature**: empty base URL makes Publish a no-op success without HTTP

```
# disabled publisher
NewPublisher("") -> Publish -> nil error, zero HTTP traffic
```

## Steps

1. Set `req.BaseURL` to empty string.
2. Do **not** start an HTTP mock.
3. Allocate an empty `HTTPCapture` so Assert can prove zero requests (no mock listens).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.BaseURL = ""
	req.Token = ""
	req.UseHTTPMock = false
	req.Capture = &HTTPCapture{}
	return nil
}
```
