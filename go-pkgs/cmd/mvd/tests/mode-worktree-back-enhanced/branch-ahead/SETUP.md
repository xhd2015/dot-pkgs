## Steps
- All tests in this branch exercise CASE B: HEAD is an ancestor of the worktree branch.
- The branch has commits not yet in main; a fast-forward merge is possible.
- The user is prompted for confirmation (Y/n, default Y).
- On confirmation: `git merge --ff-only` on main, then remove worktree and delete branch.
- On decline: abort with no changes.
- If stdin is not a TTY: fail with error.

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("branch: ahead (ff possible)")
    return nil
}
```
