# Scenario

**Feature**: until excludes the 19:00 boundary fire

```
every-5m-until-19h00m: last fire 18:55; next after that expires
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
	req.Anchor = "2026-08-24T18:50:00Z"
	req.From = "2026-08-24T18:55:00.000000001Z"
	return nil
}
```
