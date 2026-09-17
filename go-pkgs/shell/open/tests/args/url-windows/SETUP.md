# Scenario

**Feature**: windows opens a URL via `start` with an explicit window title

```
URLArgs(url, "windows") -> ["cmd", "/c", "start", "", url]
```

## Steps

1. Set `Kind=url`, `GOOS=windows`, and a URL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "url"
	req.GOOS = "windows"
	req.URL = "https://example.com/prd"
	return nil
}
```
