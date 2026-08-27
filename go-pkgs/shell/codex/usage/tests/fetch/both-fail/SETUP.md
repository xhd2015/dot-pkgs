# Scenario

**Feature**: both Codex and WHAM usage endpoints fail

```
GetJSON(*)=HTTP 403 -> Fetch error
```

## Steps

1. Set `FetchMode=both-fail`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "both-fail"
	return nil
}
```
