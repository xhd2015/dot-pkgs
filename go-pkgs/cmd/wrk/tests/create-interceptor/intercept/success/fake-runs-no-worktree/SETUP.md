# Scenario

**Feature**: enabled interceptor runs fake tool; outer creates no worktree

```
config enabled argv=[kool space create --work-dir ${work_dir}]
myrepo -> wrk -> exec kool; stdout "intercepted\n"; worktrees/ empty
```

## Steps

1. Grouping wrote simple enabled config and installed fake `kool` (exit 0).
2. Run bare `wrk` create from `myrepo`.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
