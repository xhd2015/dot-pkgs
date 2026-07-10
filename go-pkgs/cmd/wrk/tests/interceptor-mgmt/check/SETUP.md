# Scenario

**Feature**: wrk --interceptor --check validates interceptor config

```
wrk --interceptor --check
  -> absent OR valid: exit 0; empty stdout
  -> present but invalid (unknown var, empty argv when enabled, …): non-zero; stderr
```

## Steps

- Leaves seed valid/invalid/absent configs; run `--check`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--check")
	return nil
}
```
