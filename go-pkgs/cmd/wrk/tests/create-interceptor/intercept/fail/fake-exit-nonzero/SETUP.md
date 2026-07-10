# Scenario

**Feature**: interceptor child exit code propagates to wrk

```
fake kool exit 3 -> wrk exit 3; no worktree
```

## Steps

1. Reinstall fake `kool` with exit code 3 (overrides grouping install).
2. Keep enabled simple config.
3. Run bare create.

```go
func Setup(t *testing.T, req *Request) error {
	// Re-install with non-zero exit; replace ExtraEnv entries for log/exit.
	req.ExtraEnv = nil
	req.PathPrepend = ""
	req.InterceptorLog = ""
	installFakeInterceptor(t, req, fakeInterceptorName, 3)
	return nil
}
```
