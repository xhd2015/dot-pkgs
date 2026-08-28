# Scenario

**Feature**: Monthly uncapped falls back to weekly credits

```
GetJSON(monthly)=uncapped + GetJSON(credits)=weekly 2% -> PeriodType=weekly UsedPercent=2
```

## Steps

1. Set `FetchMode=monthly-open-weekly`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "monthly-open-weekly"
	return nil
}
```
