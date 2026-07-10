# Scenario

**Feature**: --check succeeds when interceptor is absent

```
no config.json -> wrk --interceptor --check -> exit 0; empty stdout
```

## Steps

1. Leave config absent.
2. Run `wrk --interceptor --check`.

```go
func Setup(t *testing.T, req *Request) error {
	assertMgmtConfigAbsent(t, req.WrkHome)
	return nil
}
```
