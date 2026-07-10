# Scenario

**Feature**: status reports enabled when enabled:true

```
create.interceptor enabled:true argv[0]=kool
wrk --interceptor --status -> state: enabled; argv0: kool
```

## Steps

1. Seed config with `enabled: true` and `argv: ["kool", "space", "create"]`.
2. Run `wrk --interceptor --status`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"kool", "space", "create"}, nil)
	return nil
}
```
