# Scenario

**Feature**: IncludeArchived true omits `--no-archived` flag

```
# include archived
IncludeArchived=true -> gh repo list without --no-archived
```

## Steps

1. Set `req.IncludeArchived` to true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.IncludeArchived = true
	return nil
}```