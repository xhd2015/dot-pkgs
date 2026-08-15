# Scenario

**Feature**: Scan from inside a checkout must discover nested git repos under it

```
# scan root is a linked worktree checkout (not a parent workspace)
# nested linked wts (external/) and nested main repos (vendor/) must appear as rows
```

## Preconditions

- Discovery only — `ListRemotes` and `ListWorktrees` remain false.
- Fake `.git` fixtures suffice; real `git` is not required.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.ListRemotes = false
	req.ListWorktrees = false
	return nil
}
```