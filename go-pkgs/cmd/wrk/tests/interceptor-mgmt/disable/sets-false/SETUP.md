# Scenario

**Feature**: --disable flips enabled true → false (create stays native)

```
enabled:true argv=[kool,…] -> wrk --interceptor --disable
  -> enabled:false; empty stdout; create would use native path
```

## Steps

1. Seed enabled interceptor.
2. Run `wrk --interceptor --disable`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"kool", "space", "create"}, nil)
	return nil
}
```
