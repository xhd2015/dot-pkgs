# Scenario

**Feature**: Codex usage fails; WHAM usage succeeds

```
GetJSON(codex/usage)=fail -> GetJSON(wham/usage)=ok -> Source=wham/usage
```

## Steps

1. Set `FetchMode=wham-fallback`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.FetchMode = "wham-fallback"
	return nil
}
```
