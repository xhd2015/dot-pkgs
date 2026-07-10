# Scenario

**Feature**: --enable flips enabled to true on existing block

```
seeded enabled:false -> wrk --interceptor --enable -> enabled:true; empty stdout
```

## Steps

1. Seed disabled interceptor with non-empty argv (post-init shape).
2. Run `wrk --interceptor --enable`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, false, []string{"echo", "wrk-interceptor-not-configured"}, nil)
	return nil
}
```
