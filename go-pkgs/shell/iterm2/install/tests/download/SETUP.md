# Scenario

**Feature**: `Download` writes zip bytes via injectable HTTP

```
Download(url, dest, HTTPClient) -> dest file or error
```

## Steps

1. Set `Operation=download`.
2. Leaves set `HTTPMode` for happy / 404 / empty-body.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "download"
	return nil
}
```
