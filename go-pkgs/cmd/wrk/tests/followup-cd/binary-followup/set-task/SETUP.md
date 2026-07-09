# Scenario

**Feature**: --set-task follow-up only when path actually moves

```
wrk --set-task <desc> + WRK_FOLLOWUP_FILE
  -> move: cd <newPath>; unchanged: empty
```

## Steps

1. Descendants create task worktree and set SetTaskDesc / CLIArgs.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "binary")
	return nil
}
```
