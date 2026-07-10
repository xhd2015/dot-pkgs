# Scenario

**Feature**: --check fails for unknown template variable

```
enabled argv with ${no_such} -> wrk --interceptor --check -> non-zero; stderr
```

## Steps

1. Seed enabled interceptor with `argv: ["echo", "${no_such}"]`.
2. Run `wrk --interceptor --check`.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"echo", "${no_such}"}, nil)
	return nil
}
```
