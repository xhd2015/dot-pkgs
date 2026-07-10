# Scenario

**Feature**: `wrk --status` unaffected by enabled create interceptor

```
config enabled + kool on PATH
myrepo -> wrk --status -> status stdout; kool not exec'd; no worktree
```

## Steps

1. Initialize clean `myrepo`.
2. Write enabled interceptor config.
3. Install fake `kool`.
4. Run `wrk --status` from the repo root.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForInterceptor(t, req)
	writeEnabledSimpleInterceptor(t, req.WrkHome)
	installFakeInterceptor(t, req, fakeInterceptorName, 0)
	req.Args = []string{"--status"}
	return nil
}
```
