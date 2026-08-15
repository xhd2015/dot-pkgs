# Scenario

**Feature**: IncludeForks false adds `--source` to gh

```
# fork filter
IncludeForks=false -> gh ... --source
```

## Steps

1. Set `req.IncludeForks` to false.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.IncludeForks = false
	return nil
}```