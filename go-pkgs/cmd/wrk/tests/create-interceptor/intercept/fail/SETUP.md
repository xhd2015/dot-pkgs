# Scenario

**Feature**: interceptor failure paths — fail hard, no native fallback

```
child exit ≠ 0 | expand error | binary missing
  -> wrk non-zero; no worktree under WRK_HOME/worktrees
```

## Steps

- Leaves reconfigure fake exit code, config templates, or PATH as needed.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
