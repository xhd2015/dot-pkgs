# Scenario

**Feature**: unknown shell — try bash then zsh on error or empty dump

```
DetectShell -> "" | other
  bash nonempty -> shell=bash (zsh not called)
  bash empty/err -> try zsh
  both fail -> error
```

## Steps

1. Default DetectShellResult to empty string (unknown).
2. Leaves may set DetectShellResult to non-bash/zsh (e.g. fish).
3. Leaves set bash/zsh stdout and fail flags.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DetectShellResult = ""
	return nil
}
```
