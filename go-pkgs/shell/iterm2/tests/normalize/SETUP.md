# Scenario

**Feature**: existing paths normalize via EvalSymlinks before script build

```
symlink dir -> OpenConfig -> EvalSymlinks -> targetDir in AppleScript
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "open-config"
	return nil
}
```