# Scenario

**Feature**: DetectShell returns zsh — dump zsh only

```
DetectShell -> "zsh"
RunLogin("zsh", env -0, …) -> dump
ResolveLoginEnvs -> shell="zsh", envs
```

## Steps

1. Set `DetectShellResult=zsh`.
2. Leaves inject zsh dump.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DetectShellResult = "zsh"
	return nil
}
```
