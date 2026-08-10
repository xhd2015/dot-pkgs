# Scenario

**Feature**: bash login shell returns path first

```
RunLogin("bash", command -v mytool) -> "/login/bash/mytool\n"
  -> Path trimmed, Via=login_shell:bash
```

## Steps

1. Set `LoginStdout["bash"]` to a synthetic absolute path (with trailing newline).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	p := filepath.Join(req.WorkDir, "login", "bash", "mytool")
	req.LoginStdout = map[string]string{
		"bash": p + "\n",
	}
	return nil
}
```
