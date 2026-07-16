# Scenario

**Feature**: CASE B — worktree branch ahead of main (fast-forward merge possible)

```
# HEAD is ancestor of worktree branch; default auto-yes merges without Proceed?
mvd --back wt → auto-yes → ff-merge + remove wt + delete branch
# --confirm restores FormatPlanPrompt [Y/n]
mvd --back --confirm [--confirm-from-stdin] wt → Proceed? → merge or abort
```

## Steps
- All tests in this branch exercise CASE B: HEAD is an ancestor of the worktree branch.
- The branch has commits not yet in main; a fast-forward merge is possible.
- Default auto-yes accepts the merge-back plan without printing `Proceed?`.
- With `--confirm`: prompt lists concrete `git -C` commands; Y/Enter merges, `n` aborts.
- `--confirm` + `--confirm-from-stdin` reads y/n from a pipe (non-TTY).
- `prompt-shows-commands`: decline under `--confirm` verifies FormatPlanPrompt output.
- Non-TTY / piped stdin without `--confirm` still succeed (auto-yes).

```go
func Setup(t *testing.T, req *Request) error {
	t.Logf("branch: ahead (ff possible)")
	return nil
}
```
