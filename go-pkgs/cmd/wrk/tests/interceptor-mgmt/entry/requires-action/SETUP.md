# Scenario

**Feature**: bare --interceptor without action errors

```
wrk --interceptor -> non-zero; stderr mentions action or usage
```

## Steps

1. Run `wrk --interceptor` with no action flags.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--interceptor"}
	return nil
}
```
