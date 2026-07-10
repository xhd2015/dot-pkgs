# Scenario

**Feature**: wrk --interceptor --init writes disabled neutral stub

```
wrk --interceptor --init [--force]
  -> no prior interceptor: write enabled:false + stub argv
  -> existing without --force: refuse (non-zero), file unchanged
  -> existing with --force: replace interceptor block only
  -> existing config without interceptor: merge stub; preserve other keys
```

## Steps

- Leaves seed config variants; run `--init` or `--init --force`.

## Context

- Stub argv starts with `echo` / `wrk-interceptor-not-configured`.
- Empty stdout preferred on success.

```go
func Setup(t *testing.T, req *Request) error {
	// Leaf overrides Args for --force cases.
	if len(req.Args) == 0 {
		req.Args = interceptorMgmtArgs("--init")
	}
	return nil
}
```
