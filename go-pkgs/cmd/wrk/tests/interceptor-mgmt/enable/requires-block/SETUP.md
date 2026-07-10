# Scenario

**Feature**: --enable fails when no interceptor block exists

```
no create.interceptor -> wrk --interceptor --enable -> non-zero; hint --init
```

## Steps

1. Leave config absent.
2. Run `wrk --interceptor --enable`.

```go
func Setup(t *testing.T, req *Request) error {
	assertMgmtConfigAbsent(t, req.WrkHome)
	return nil
}
```
