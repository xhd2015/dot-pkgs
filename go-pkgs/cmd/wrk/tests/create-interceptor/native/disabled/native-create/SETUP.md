# Scenario

**Feature**: native create when interceptor is disabled

```
enabled:false + kool on PATH
myrepo -> wrk -> native worktree; interceptor log empty
```

## Steps

1. Initialize `myrepo`.
2. Write config with `enabled: false` and `argv: ["kool", "should-not-run"]`.
3. Install fake `kool` that would log argv if invoked.
4. Run bare `wrk` create.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForInterceptor(t, req)
	writeInterceptorConfig(t, req.WrkHome, false, []string{fakeInterceptorName, "should-not-run"}, nil)
	installFakeInterceptor(t, req, fakeInterceptorName, 0)
	return nil
}
```
