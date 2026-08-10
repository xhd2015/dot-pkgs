# Scenario

**Feature**: successful Publish against an HTTP mock (2xx)

```
# mock hub accepts publish
httptest.Server 200 <- Publisher.Publish POST /publish JSON
```

## Steps

1. Start an HTTP mock that returns 200 and records requests into `req.Capture`.
2. Set `req.BaseURL` to the mock server URL.
3. Leaves set Token (empty vs non-empty).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UseHTTPMock = true
	req.MockStatusCode = 200
	req.Capture = &HTTPCapture{}
	srv := startPublishMock(t, req.MockStatusCode, req.Capture)
	req.BaseURL = srv.URL
	return nil
}
```
