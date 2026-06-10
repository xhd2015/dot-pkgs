## Preconditions

- A TCP listener can be started on a random port to simulate an upstream proxy

## Steps

- Testing `--upstream-proxy` behavior: accessible vs unreachable at startup, with and without `--fallback-direct`

```go
func Setup(t *testing.T, req *Request) error {
	t.Log("entering upstream proxy test mode")
	return nil
}
```
