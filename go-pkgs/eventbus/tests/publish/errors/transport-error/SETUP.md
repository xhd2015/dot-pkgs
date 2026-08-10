# Scenario

**Feature**: transport failure (closed server) causes Publish to return an error

```
# network / dial failure
start mock -> close immediately -> Publish(baseURL) -> non-nil error
```

## Steps

1. Start an HTTP mock, capture its URL, then close it before `Run`.
2. Leave `req.BaseURL` pointing at the closed listener so dial/connect fails.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.UseHTTPMock = true
	req.CloseMockEarly = true
	if req.Capture == nil {
		req.Capture = &HTTPCapture{}
	}
	srv := startPublishMock(t, 200, req.Capture)
	req.BaseURL = srv.URL
	srv.Close() // close now; t.Cleanup Close is harmless after
	return nil
}
```
