# Scenario

**Feature**: --init --force replaces existing interceptor with stub

```
existing argv=[custom-tool] -> wrk --interceptor --init --force
  -> stub enabled:false; custom argv gone; exit 0
```

## Steps

1. Seed custom interceptor.
2. Run `wrk --interceptor --init --force`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--init", "--force")
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"custom-tool", "replace-me"}, nil)
	return nil
}
```
