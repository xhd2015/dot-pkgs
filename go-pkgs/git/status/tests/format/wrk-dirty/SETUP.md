# Scenario

**Feature**: full wrk counts format as four-segment dirty string

```
{1,1,1,1} -> FormatWrk -> dirty (1 added, 1 changed, 1 renamed, 1 deleted)
```

## Steps

1. Set `req.Op` to `"format-wrk"`.
2. Set one count in each wrk bucket.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, req *Request) error {
	req.Op = "format-wrk"
	req.WrkCounts = status.WrkCounts{Added: 1, Changed: 1, Renamed: 1, Deleted: 1}
	return nil
}
```