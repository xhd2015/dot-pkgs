## Preconditions
- Git must be available (all tests involve git repos and worktrees).

## Steps
- All tests in this mode exercise safety scenarios where plain moves and worktree operations involve overlapping paths.
- Tests cover: parent directory moves containing tracked repos, stale MainRepo in history, nesting worktrees, and long-chain --back resolution.
- Some tests document currently-broken behavior (stale worktree .git files after parent moves, position mismatch blocking --back on non-last worktree).

```go
func Setup(t *testing.T, req *Request) error {
    skipIfNoGit(t)
    t.Logf("mode: safety")
    return nil
}
```
