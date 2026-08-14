# Scenario

**Feature**: positional SHA does not resolve in the repo

```
RunCLI -C <repo> not-a-real-sha -m x -> Error: unknown revision: not-a-real-sha
```

## Steps

1. Create a real git repo with one commit (so the dir is a repo).
2. Pass a token that is not a revision.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = initRepo(t)
	commitFile(t, req.Dir, "README", "init\n", "init")
	req.Args = []string{"-C", req.Dir, "not-a-real-sha", "-m", "x"}
	return nil
}
```
