# Scenario

**Feature**: interceptor succeeds (exit 0) — outer skips native create side effects

```
kool exit 0 -> wrk exit 0; stdout from fake; no WRK_HOME worktree from outer
```

## Steps

- Inherit enabled config + fake with exit 0.
- Leaves assert expand details, follow-up, or auto-record as needed.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
