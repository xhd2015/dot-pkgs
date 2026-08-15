# Scenario

**Feature**: zero wrk counts format as clean

```
empty WrkCounts -> FormatWrk -> "clean"
```

## Steps

1. Set `req.Op` to `"format-wrk"`.
2. Leave `req.WrkCounts` at zero value.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "format-wrk"
	req.WrkCounts = status.WrkCounts{}
	return nil
}
```