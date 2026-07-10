# Scenario

**Feature**: --path prints config path when file is missing

```
no config.json
wrk --interceptor --path -> abs({WRK_HOME}/config.json) + \n; exit 0
```

## Steps

1. Do not write `config.json`.
2. Run `wrk --interceptor --path` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	// Parent sets --path; ensure file is absent.
	assertMgmtConfigAbsent(t, req.WrkHome)
	return nil
}
```
