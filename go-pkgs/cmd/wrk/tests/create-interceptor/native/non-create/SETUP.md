# Scenario

**Feature**: non-create modes ignore create.interceptor

```
config enabled + fake on PATH
wrk --status (or other non-create) -> native status; fake not exec'd
```

## Steps

- Enabled interceptor + fake installed.
- Leaf runs a non-create mode (`--status`).

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
