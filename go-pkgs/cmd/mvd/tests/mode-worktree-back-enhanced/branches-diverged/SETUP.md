# Scenario

**Feature**: CASE C — worktree branch diverged from main (rebase then merge)

```
# neither branch is ancestor of the other; default auto-yes rebases without Proceed?
mvd --back wt → auto-yes → rebase+merge + remove wt + delete branch
# --confirm restores FormatPlanPrompt [Y/n]
mvd --back --confirm [--confirm-from-stdin] wt → Proceed? → rebase+merge or abort
```

## Steps
- All tests in this branch exercise CASE C: neither HEAD nor the worktree branch is an ancestor of the other.
- Default auto-yes accepts the rebase+merge plan without printing `Proceed?`.
- With `--confirm`: prompt lists concrete `git -C` commands; Y/Enter rebases then ff-merges; `n` aborts.
- On confirmation: worktree branch is rebased onto main HEAD; on success `git merge --ff-only` on main, remove worktree, delete branch.
- If the rebase conflicts: abort the rebase and report error.
- `prompt-shows-commands`: decline under `--confirm` verifies FormatPlanPrompt output.
- Non-TTY without flags succeeds (auto-yes), not a hard confirmation error.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Logf("branches: diverged")
	return nil
}
```
