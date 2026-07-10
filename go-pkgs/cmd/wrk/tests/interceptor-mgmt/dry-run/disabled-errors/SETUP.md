# Scenario

**Feature**: --dry-run errors when interceptor is disabled

```
enabled:false -> wrk --interceptor --dry-run -> non-zero (must be enabled)
```

## Steps

1. Seed disabled interceptor with non-empty argv.
2. Run `wrk --interceptor --dry-run`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, false, []string{"echo", "noop"}, nil)
	req.Args = interceptorMgmtArgs("--dry-run")
	return nil
}
```
