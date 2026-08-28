# Scenario

**Feature**: ForceRefresh option is passed to Ensure

```
ForceRefresh=true -> Ensure(ForceRefresh=true) -> billing OK
```

## Steps

1. Set `FetchMode=force-refresh`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "force-refresh"
	return nil
}
```
