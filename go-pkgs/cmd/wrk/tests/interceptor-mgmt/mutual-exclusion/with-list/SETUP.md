# Scenario

**Feature**: --interceptor cannot combine with --list

```
wrk --interceptor --status --list -> non-zero; mutual exclusion error
```

## Steps

1. Run `wrk --interceptor --status --list` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--interceptor", "--status", "--list"}
	return nil
}
```
