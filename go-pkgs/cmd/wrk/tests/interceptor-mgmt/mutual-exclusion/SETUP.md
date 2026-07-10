# Scenario

**Feature**: --interceptor is mutually exclusive with other wrk modes

```
wrk --interceptor … + --list / other modes -> non-zero; clear mutual exclusion error
```

## Steps

- Leaves combine interceptor management with another mode flag.

```go
func Setup(t *testing.T, req *Request) error {
	// Marker: mutual-exclusion leaves append a second mode flag to Args.
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
