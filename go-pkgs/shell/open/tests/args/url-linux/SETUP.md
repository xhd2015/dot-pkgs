# Scenario

**Feature**: linux opens a URL with xdg-open

```
URLArgs(url, "linux") -> ["xdg-open", url]
```

## Steps

1. Set `Kind=url`, `GOOS=linux`, and a URL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "url"
	req.GOOS = "linux"
	req.URL = "https://example.com/prd"
	return nil
}
```
