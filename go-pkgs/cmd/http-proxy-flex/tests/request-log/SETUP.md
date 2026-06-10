## Preconditions

- A TCP listener can simulate an upstream proxy
- A local HTTP server can serve as request target

## Steps

- Testing that each proxied request logs whether it goes through upstream proxy or direct

```go
func Setup(t *testing.T, req *Request) error {
	t.Log("entering request log test mode")
	return nil
}
```
