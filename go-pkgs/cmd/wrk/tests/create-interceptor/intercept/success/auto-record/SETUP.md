# Scenario

**Feature**: auto-record still runs before interceptor exec

```
create intercept -> projects.json records main repo (source auto)
                 -> events.jsonl command remains "create"
```

## Steps

- Success intercept path; assert projects.json and events.jsonl.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
