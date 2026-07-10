# Scenario

**Feature**: --init merges stub into existing config without wiping unrelated keys

```
config.json has version + unrelated top-level key, no interceptor
wrk --interceptor --init -> adds create.interceptor stub; preserves other keys
```

## Steps

1. Write config with `"version": 1` and `"notes": "keep-me"` (no create.interceptor).
2. Run `wrk --interceptor --init`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--init")
	writeMgmtRawConfig(t, req.WrkHome, `{
  "version": 1,
  "notes": "keep-me"
}
`)
	return nil
}
```
