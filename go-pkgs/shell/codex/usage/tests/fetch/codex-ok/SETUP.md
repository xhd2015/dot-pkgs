# Scenario

**Feature**: Codex usage endpoint succeeds

```
GetJSON(codex/usage)=200 spend_control -> Source=codex/usage remaining=38
```

## Steps

1. Set `FetchMode=codex-ok`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "codex-ok"
	return nil
}
```
