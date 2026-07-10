# Scenario

**Feature**: status reports state absent with no interceptor

```
no config.json (or no create.interceptor)
wrk --interceptor --status -> state: absent; argv0: -
```

## Steps

1. Leave `config.json` missing.
2. Run `wrk --interceptor --status`.

```go
func Setup(t *testing.T, req *Request) error {
	assertMgmtConfigAbsent(t, req.WrkHome)
	return nil
}
```
