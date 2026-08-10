# Scenario

**Feature**: non-empty TabSpec.Cwd emits cd; empty Cwd does not force project cd

```
Tab A Cwd=/tmp/iterm2-tabset-cwd + Tab B Cwd=""
  -> BuildTabSetNewWindowScript
  -> only tab A has a cd write for its path; both commands present
```

## Steps

1. Tab `with-cwd`: Cwd `/tmp/iterm2-tabset-cwd`, command `echo-with-cwd`.
2. Tab `no-cwd`: empty Cwd, command `echo-no-cwd`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {

	req.TabSetName = "cwd-mix"
	req.Tabs = []TabSpecInput{
		{ID: "with-cwd", Name: "WithCwd", Command: "echo-with-cwd", Cwd: "/tmp/iterm2-tabset-cwd"},
		{ID: "no-cwd", Name: "NoCwd", Command: "echo-no-cwd", Cwd: ""},
	}
	return nil
}
```
