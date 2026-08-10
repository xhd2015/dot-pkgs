# Scenario

**Feature**: Publisher.Publish posts Event JSON to the hub publish endpoint

```
# HTTP publish path
Publisher(baseURL, token?) -> Publish(ctx, Event) -> POST {baseURL}/publish
empty baseURL -> no-op success
non-2xx / transport fail -> error
```

## Steps

1. Set `req.Op` to `"publish"`.
2. Leaves configure BaseURL / Token / HTTP mock.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "publish"
	return nil
}
```
