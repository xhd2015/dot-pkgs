# Scenario

**Feature**: a URL runs the default handler once

```
URLConfig(url, cfg) -> Run(["open", url]) -> Result
```

## Steps

1. Set `Kind=url`, `GOOS=darwin`, and a URL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "url"
	req.GOOS = "darwin"
	req.URL = "https://example.com/prd"
	return nil
}
```
