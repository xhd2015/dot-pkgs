# Scenario

**Feature**: intercept create still auto-records the project

```
myrepo -> wrk (intercept) -> projects.json has myrepo source=auto
                          -> events.jsonl command=create exit 0
```

## Steps

1. Grouping fixtures (repo + enabled interceptor + fake).
2. Bare create from main repo cwd.

```go
func Setup(t *testing.T, req *Request) error {
	// Grouping marker: keep helpers linked for descendant leaves.
	ensureInterceptorHelpersUsed()
	return nil
}
```
