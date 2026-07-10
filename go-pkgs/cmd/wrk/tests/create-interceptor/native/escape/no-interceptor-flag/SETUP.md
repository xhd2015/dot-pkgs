# Scenario

**Feature**: `--no-interceptor` skips enabled create interceptor

```
config enabled + kool on PATH
wrk --no-interceptor -> native worktree; kool not exec'd
```

## Steps

1. Grouping already wrote enabled config and installed fake.
2. Pass `--no-interceptor` on the create invocation.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--no-interceptor"}
	return nil
}
```
