# Scenario

**Feature**: bash login resolves remaining name; From=bash

```
RunLogin("bash", … mytool …) -> "/login/bash/mytool\n"
  -> Item Path trimmed, From="bash"
```

## Steps

1. Single name `mytool`.
2. Set LoginStdout["bash"] to a synthetic absolute path (trailing newline OK).

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"mytool"}
	p := filepath.Join(req.WorkDir, "login", "bash", "mytool")
	req.LoginStdout = map[string]string{
		"bash": p + "\n",
	}
	return nil
}
```
