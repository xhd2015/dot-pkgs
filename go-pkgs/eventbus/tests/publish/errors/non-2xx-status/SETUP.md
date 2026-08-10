# Scenario

**Feature**: non-2xx HTTP status causes Publish to return an error

```
# hub rejects
httptest.Server 500 <- Publish -> non-nil error
```

## Steps

1. Start HTTP mock that responds with status 500.
2. Point `req.BaseURL` at the mock.

```go
import (
	"net/http"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UseHTTPMock = true
	req.MockStatusCode = http.StatusInternalServerError
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	srv := startPublishMock(t, req.MockStatusCode, req.Capture)
	req.BaseURL = srv.URL
	return nil
}
```
