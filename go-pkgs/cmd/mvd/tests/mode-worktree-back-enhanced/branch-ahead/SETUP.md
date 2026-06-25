# Scenario

**Feature**: CASE B — worktree branch ahead of main (fast-forward merge possible)

```
# HEAD is ancestor of worktree branch; confirmation before merge-back
mvd --back wt → FormatPlanPrompt (git -C commands) → [Y/n] → merge or abort
```

## Steps
- All tests in this branch exercise CASE B: HEAD is an ancestor of the worktree branch.
- The branch has commits not yet in main; a fast-forward merge is possible.
- The user is prompted for confirmation (Y/n, default Y) with concrete `git -C` commands listed.
- On confirmation: `git merge --ff-only` on main, then remove worktree and delete branch.
- On decline: abort with no changes.
- `prompt-shows-commands`: decline path verifies FormatPlanPrompt output before abort.
- If stdin is not a TTY: fail with error.

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("branch: ahead (ff possible)")
    return nil
}
```
