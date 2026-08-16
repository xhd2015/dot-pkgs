# Scenario

**Feature**: `ResolveLoginEnvs` dumps login environ for the detected shell

```
DetectShell() -> bash|zsh|unknown
  bash/zsh -> dump that shell only
  unknown  -> try bash then zsh on error/empty
```

## Steps

1. Set `Op=resolve`.
2. Child groups set DetectShellResult; leaves set dumps / fails.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "resolve"
	return nil
}
```
