# Scenario

**Feature**: --show prints interceptor object when present

```
seeded create.interceptor -> wrk --interceptor --show -> pretty JSON; exit 0
```

## Steps

1. Seed enabled interceptor with known argv.
2. Run `wrk --interceptor --show`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"kool", "space", "create"}, nil)
	return nil
}
```
