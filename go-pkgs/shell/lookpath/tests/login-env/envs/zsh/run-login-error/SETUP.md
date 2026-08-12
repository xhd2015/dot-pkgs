# Scenario

**Feature**: zsh Envs surfaces RunLogin failure as error

```
RunLogin("zsh", …) -> error
ResolveZshLoginEnvs -> non-nil error
```

## Steps

1. Set LoginFail=true (no successful dump).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.LoginFail = true
	return nil
}
```
