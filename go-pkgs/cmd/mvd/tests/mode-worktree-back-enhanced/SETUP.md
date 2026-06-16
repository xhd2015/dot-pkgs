## Preconditions
- Git must be available.
- The `script` command must be available (for TTY simulation tests).

## Steps
- All tests in this mode exercise `mvd --back` on a git worktree entry.
- The enhanced behavior adds CASE B (branch ahead, ff merge with prompt) and CASE C (branches diverged, rebase) beyond the existing merged/unmerged/dirty checks.
- TTY simulation uses `script -q /dev/null` to create a pseudo-terminal for the child process.

```go
func Setup(t *testing.T, req *Request) error {
    skipIfNoGit(t)
    t.Logf("mode: worktree-back-enhanced")
    return nil
}
```
