# Scenario

**Feature**: bash RunLogin failure is soft — continue to zsh hit

```
RunLogin("bash", …) -> error (soft continue)
RunLogin("zsh", …)  <- GOPATH=/tmp/from-zsh\0
ResolveGoPathWith -> ("/tmp/from-zsh", nil)
```

## Steps

1. Set BashFail=true (login shell failure must not abort cascade).
2. Zsh dump still provides GOPATH (parent group).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashFail = true
	return nil
}
```
