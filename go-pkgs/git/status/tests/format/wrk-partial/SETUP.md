# Scenario

**Feature**: partial wrk counts still emit all five dirty segments

```
{Changed:1} -> FormatWrk -> dirty (0 staged, 1 changed, 0 renamed, 0 deleted, 0 untracked)
```

## Steps

1. Set `req.Op` to `"format-wrk"`.
2. Set only `Changed` to 1.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Op = "format-wrk"
	req.WrkCounts = status.WrkCounts{Changed: 1}
	return nil
}
```
