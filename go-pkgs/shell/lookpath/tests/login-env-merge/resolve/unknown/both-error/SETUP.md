# Scenario

**Feature**: unknown detect — both bash and zsh RunLogin fail → error

```
DetectShell -> ""
RunLogin("bash") -> error
RunLogin("zsh")  -> error
ResolveLoginEnvs -> error
```

## Steps

1. BashFail=true and ZshFail=true.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BashFail = true
	req.ZshFail = true
	return nil
}
```
