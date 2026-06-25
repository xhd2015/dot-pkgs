# Scenario

**Feature**: CASE C — worktree branch diverged from main (rebase then merge)

```
# neither branch is ancestor of the other; confirmation before rebase+merge
mvd --back wt → FormatPlanPrompt (git -C commands) → [Y/n] → rebase+merge or abort
```

## Steps
- All tests in this branch exercise CASE C: neither HEAD nor the worktree branch is an ancestor of the other.
- The user is prompted for confirmation (Y/n, default Y) with concrete `git -C` commands listed.
- On confirmation: the worktree branch is rebased onto main HEAD.
- If the rebase succeeds: `git merge --ff-only` on main, remove worktree, delete branch.
- If the rebase conflicts: abort the rebase and report error.
- On decline: abort with no changes.
- `prompt-shows-commands`: decline path verifies FormatPlanPrompt output before abort.
- If stdin is not a TTY: fail with error.

```go
func Setup(t *testing.T, req *Request) error {
    t.Logf("branches: diverged")
    return nil
}
```
