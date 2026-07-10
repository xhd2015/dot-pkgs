# Scenario

**Feature**: escape hatches force native create despite enabled interceptor

```
config enabled + fake on PATH
  + --no-interceptor  OR  WRK_NO_INTERCEPTOR=1
  -> native worktree create; fake not invoked
```

## Steps

- Write enabled interceptor config and install fake `kool`.
- Leaves apply either CLI `--no-interceptor` or env `WRK_NO_INTERCEPTOR=1`.

```go
func Setup(t *testing.T, req *Request) error {
	setupMainRepoForInterceptor(t, req)
	writeEnabledSimpleInterceptor(t, req.WrkHome)
	installFakeInterceptor(t, req, fakeInterceptorName, 0)
	return nil
}
```
