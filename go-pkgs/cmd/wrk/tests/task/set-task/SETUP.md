# Scenario

**Feature**: wrk --set-task renames a linked worktree via git worktree move

```
# inside linked worktree, parse branch name, compute new slug, warn+move if TTY
wrk --set-task "new desc" -> parse {branchBase}-{YYYY-MM-DD}[-slug][-N]
                         -> git worktree move (TTY required)
```

## Preconditions

- Must be run from inside a linked worktree (`.git` is a file).
- The worktree branch must follow the wrk naming pattern.

## Steps

- Create a worktree (with or without --task).
- Run `wrk --set-task <desc>` from inside the worktree.
- Non-TTY environment → error; TTY → confirm then move.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

```