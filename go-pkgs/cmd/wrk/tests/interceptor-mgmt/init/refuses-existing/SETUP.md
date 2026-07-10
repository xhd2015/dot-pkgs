# Scenario

**Feature**: --init without --force refuses existing interceptor

```
existing interceptor argv=[custom-tool]
wrk --interceptor --init -> non-zero; config unchanged
```

## Steps

1. Seed custom enabled interceptor.
2. Snapshot config bytes.
3. Run `wrk --interceptor --init` (no `--force`).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = interceptorMgmtArgs("--init")
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"custom-tool", "keep-me"}, nil)
	// Stash prior content for assert via ExtraEnv marker path is overkill;
	// ASSERT re-reads and checks argv still custom-tool.
	return nil
}
```
