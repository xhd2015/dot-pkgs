# Scenario

**Feature**: --dry-run errors when interceptor is absent

```
no create.interceptor -> wrk --interceptor --dry-run -> non-zero
```

## Steps

1. Leave config absent.
2. Run `wrk --interceptor --dry-run`.

```go
func Setup(t *testing.T, req *Request) error {
	assertMgmtConfigAbsent(t, req.WrkHome)
	req.Args = interceptorMgmtArgs("--dry-run")
	return nil
}
```
