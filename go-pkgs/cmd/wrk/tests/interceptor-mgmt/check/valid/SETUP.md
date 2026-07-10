# Scenario

**Feature**: --check succeeds for valid enabled interceptor

```
enabled argv without unknown vars -> wrk --interceptor --check -> exit 0; empty stdout
```

## Steps

1. Seed enabled interceptor with simple argv (no templates).
2. Run `wrk --interceptor --check`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"echo", "ok"}, nil)
	return nil
}
```
