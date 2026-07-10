# Scenario

**Feature**: --dry-run prints expanded argv including task builtin

```
enabled argv=["echo","task=${task}"]
wrk --interceptor --dry-run -- -t hi
  -> stdout lines: echo\ntask=hi\n; no interceptor exec
```

## Steps

1. Seed enabled interceptor with `argv: ["echo", "task=${task}"]`.
2. Run `wrk --interceptor --dry-run -- -t hi` from neutral (non-git) cwd.
3. Assert expanded lines; process must not require a successful worktree create.

```go
func Setup(t *testing.T, req *Request) error {
	writeMgmtInterceptorConfig(t, req.WrkHome, true, []string{"echo", "task=${task}"}, nil)
	req.Args = []string{"--interceptor", "--dry-run", "--", "-t", "hi"}
	return nil
}
```
