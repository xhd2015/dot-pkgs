# Scenario

**Feature**: --init creates disabled stub when no config exists

```
no config.json -> wrk --interceptor --init
  -> config.json with enabled:false + non-empty stub argv; exit 0
```

## Steps

1. Ensure no prior `config.json`.
2. Run `wrk --interceptor --init`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--init")
	assertMgmtConfigAbsent(t, req.WrkHome)
	return nil
}
```
