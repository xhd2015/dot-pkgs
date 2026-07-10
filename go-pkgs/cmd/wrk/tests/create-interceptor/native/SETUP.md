# Scenario

**Feature**: create interceptor inactive — native create or non-create modes

```
# intercept decision = skip
no config | enabled:false | --no-interceptor | WRK_NO_INTERCEPTOR=1 | non-create
  -> wrk native path; fake tool not exec'd
```

## Steps

- Leaves either omit config, disable interceptor, set an escape hatch, or run a non-create mode.
- When a fake is installed, asserts require empty interceptor log (not invoked).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
