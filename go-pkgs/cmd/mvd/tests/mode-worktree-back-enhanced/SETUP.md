# Scenario

**Feature**: enhanced `mvd --back` for worktree entries — CASE B (ff-merge) and CASE C (rebase)

```
# default: auto-yes merge-back plan confirm (no Proceed?); --confirm restores Y/n
mvd --back wt → auto-yes → ff-merge or rebase+merge + remove wt + delete branch
mvd --back --confirm wt → FormatPlanPrompt → [Y/n]
```

## Preconditions
- Git must be available.
- The `script` command must be available (for TTY simulation tests).

## Steps
- All tests in this mode exercise `mvd --back` on a git worktree entry.
- Enhanced behavior adds CASE B (branch ahead, ff merge) and CASE C (branches diverged, rebase) beyond merged/unmerged/dirty checks.
- Default is auto-yes for plan confirm; `--confirm` forces interactive prompts; `--confirm-from-stdin` reads y/n from a pipe when used with `--confirm`.
- TTY simulation uses `script -q /dev/null` to create a pseudo-terminal for the child process.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)
	t.Logf("mode: worktree-back-enhanced")
	return nil
}
```
