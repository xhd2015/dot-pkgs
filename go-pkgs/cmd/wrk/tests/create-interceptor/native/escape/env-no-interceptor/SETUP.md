# Scenario

**Feature**: `WRK_NO_INTERCEPTOR=1` skips enabled create interceptor

```
config enabled + kool on PATH
WRK_NO_INTERCEPTOR=1 wrk -> native worktree; kool not exec'd
```

## Steps

1. Grouping already wrote enabled config and installed fake.
2. Export `WRK_NO_INTERCEPTOR=1` via `Request.ExtraEnv`.

```go
func Setup(t *testing.T, req *Request) error {
	req.ExtraEnv = append(req.ExtraEnv, "WRK_NO_INTERCEPTOR=1")
	return nil
}
```
