# Scenario

**Feature**: starting at/after until expires immediately

```
every-5m-until-19h00m anchor=20:00 -> no fire
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
	req.Expr = "every-5m-until-19h00m"
	req.Anchor = "2026-08-24T20:00:00Z"
	req.From = "2026-08-24T20:00:00Z"
	return nil
}
```
