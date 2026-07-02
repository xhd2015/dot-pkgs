# Scenario

**Feature**: wrk --task and wrk --set-task for worktree task descriptions

```
# spawn with --task appends slug to dir/branch names
wrk --task "fix login" -> git worktree add -> stdout path + .wrk-task

# --set-task inside worktree renames via git worktree move
wrk --set-task "new desc" -> parse branch -> compute new names -> git worktree move
```

## Preconditions

- Git must be available.
- The wrk binary is built once per test session.

## Steps

- Spawn tests create a git repo and run `wrk --task <desc>` from it.
- Set-task tests run `wrk --set-task <desc>` from inside a linked worktree.
- `WRK_HOME` is isolated per test at `{WorkRoot}/.wrk`.
- `WRK_DATE` is fixed to `2026-06-30`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```
