# Scenario

**Feature**: HTTP 404 from npm latest endpoint → error

```
GET /@openai/codex/latest -> 404 -> error
```

## Steps

1. Set `HTTPMode=http-404`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.HTTPMode = "http-404"
	return nil
}
```
