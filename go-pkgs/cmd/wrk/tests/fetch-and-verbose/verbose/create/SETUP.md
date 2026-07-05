# Scenario

**Feature**: wrk -v on no-args create logs major git commands

```
no-args create -> stderr contains worktree add (and likely checkout)
minor reads (rev-parse, status) not logged
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```