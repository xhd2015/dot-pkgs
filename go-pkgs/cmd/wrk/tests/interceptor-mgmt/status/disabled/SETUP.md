# Scenario

**Feature**: status reports disabled when enabled:false

```
create.interceptor enabled:false argv[0]=echo
wrk --interceptor --status -> state: disabled; argv0: echo
```

## Steps

1. Seed config with `enabled: false` and `argv: ["echo", "noop"]`.
2. Run `wrk --interceptor --status`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, false, []string{"echo", "noop"}, nil)
	return nil
}
```
