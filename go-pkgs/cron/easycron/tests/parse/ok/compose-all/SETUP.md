# Scenario

**Feature**: Parse composes at + until + not-within

```
Parse("every-5m-at-0m-until-19h00m-not-within-19h00m-to-06h30m")
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
	req.Expr = "every-5m-at-0m-until-19h00m-not-within-19h00m-to-06h30m"
	return nil
}
```
