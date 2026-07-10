# Scenario

**Feature**: --show errors when interceptor block is absent

```
no create.interceptor -> wrk --interceptor --show -> non-zero; stderr message
```

## Steps

1. Leave `config.json` missing.
2. Run `wrk --interceptor --show`.

```go
func Setup(t *testing.T, req *Request) error {
	assertMgmtConfigAbsent(t, req.WrkHome)
	return nil
}
```
