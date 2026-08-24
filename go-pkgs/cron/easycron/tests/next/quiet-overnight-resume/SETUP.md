# Scenario

**Feature**: not-within skips evening and resumes at/after 06:30

```
every-5m-not-within-19h00m-to-06h30m from 19:01 -> >= 06:30 next day active
```

## Steps

1. Configure the request fields below.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Expr = "every-5m-not-within-19h00m-to-06h30m"
	req.Anchor = "2026-08-24T18:55:00Z"
	req.From = "2026-08-24T19:01:00Z"
	return nil
}
```
