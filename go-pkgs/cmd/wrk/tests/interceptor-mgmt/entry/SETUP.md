# Scenario

**Feature**: --interceptor parent requires exactly one action flag

```
wrk --interceptor (no action) -> non-zero; usage / action required
```

## Steps

- Leaf runs bare parent flag without action.

```go
func Setup(t *testing.T, req *Request) error {
	// Default bare parent; leaf may refine.
	req.Args = []string{"--interceptor"}
	return nil
}
```
