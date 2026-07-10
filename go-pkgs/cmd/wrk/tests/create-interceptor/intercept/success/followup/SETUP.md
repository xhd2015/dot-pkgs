# Scenario

**Feature**: intercept path skips outer home-gated follow-up `cd`

```
WRK_FOLLOWUP_FILE set + shell cwd = user home
native create would write: cd <worktree>
intercept path -> follow-up file has no cd from outer wrk
```

## Steps

- Leaves open follow-up channel and place shell cwd at FakeHome with create via dir arg.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
