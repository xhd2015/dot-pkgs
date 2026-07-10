# Scenario

**Feature**: create mode with interceptor enabled — replace native create

```
# intercept active
config enabled + create mode (no escape)
  -> expand argv/vars -> exec kool on PATH
  -> no outer worktree under WRK_HOME/worktrees
  -> wrk exit = child exit
```

## Steps

- Default grouping: init `myrepo`, write enabled simple interceptor, install fake `kool`.
- Success/fail leaves may override config (recipe, unknown var, missing binary, exit code).

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForInterceptor(t, req)
	writeEnabledSimpleInterceptor(t, req.WrkHome)
	installFakeInterceptor(t, req, fakeInterceptorName, 0)
	return nil
}
```
