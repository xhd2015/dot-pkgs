# Scenario

**Feature**: no `$WRK_HOME/config.json` — interceptor absent

```
# storage has only default WRK_HOME layout (no config.json)
myrepo -> wrk (create) -> native worktree under WRK_HOME/worktrees
```

## Steps

- Do not write config.json.
- Create from main repo cwd.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
