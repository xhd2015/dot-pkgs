## Preconditions
- Git must be available.

## Steps
- All tests in this mode exercise `mvd -w | --worktree SRC DST` to create a git worktree at DST from a git repository at SRC.
- The worktree entry is tracked in history with git metadata (main repo path and branch name).
- `mvd --back` for worktree entries uses `git worktree remove` and `git branch -D` instead of `os.Rename`.
- Worktree back is gated on clean working tree and merged branch status.
- Without the `-w` flag, mvd does a plain `os.Rename` even for worktree directories (but updates child worktree `.git` files).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    skipIfNoGit(t)
    t.Logf("mode: worktree")
    return nil
}
```
