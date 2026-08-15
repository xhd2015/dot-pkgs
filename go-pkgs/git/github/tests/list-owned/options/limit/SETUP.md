# Scenario

**Feature**: custom Limit forwards `--limit` to gh

```
# limit flag
Options.Limit=42 -> gh repo list ... --limit 42
```

## Steps

1. Set `req.Limit` to 42.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Limit = 42
	return nil
}```