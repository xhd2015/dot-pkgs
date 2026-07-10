# Scenario

**Feature**: wrk --interceptor --dry-run expands argv without exec

```
wrk --interceptor --dry-run [--] [create-args...]
  -> present+enabled: print expanded argv one element per line; no exec
  -> absent or disabled: non-zero; clear error
```

## Steps

- Leaves seed config; pass optional create args after `--`.

## Context

- Requires interceptor present and enabled.
- Builtins (`task`, `work_dir`, …) expand like create; no worktree create required.
- Management must not exec the interceptor binary.

```go
func Setup(t *testing.T, req *Request) error {
	// Default dry-run with no create args; leaves override Args and seed config.
	req.Args = interceptorMgmtArgs("--dry-run")
	return nil
}
```
