# Scenario

**Feature**: Parse every-5m-until-19h00m sets Until 19:00

```
Parse("every-5m-until-19h00m") -> Until=19:00
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
	return nil
}
```
