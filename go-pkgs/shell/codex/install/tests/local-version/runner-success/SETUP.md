# Scenario

**Feature**: injected RunVersion returns codex-cli version stdout

```
RunVersion(bin) -> "codex-cli 0.147.0" -> LocalVersion returns same raw string
```

## Steps

1. Set success runner output; LookPath hit.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.VersionCmdOutput = "codex-cli 0.147.0"
	req.VersionCmdFail = false
	req.LookPathMiss = false
	return nil
}
```
