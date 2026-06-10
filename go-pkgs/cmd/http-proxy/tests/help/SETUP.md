## Preconditions

- The `http-proxy` binary can be compiled

## Steps

- Testing the `--help` flag behavior

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
