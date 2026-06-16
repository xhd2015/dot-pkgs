## Preconditions

- The `http-proxy` binary can be compiled
- No process is listening on port 19999 (unused by other tests)

## Steps

- Testing the `--listen-port` flag: default vs custom values
- Both children use `--upstream-proxy` pointing to a dead port with `--fallback-direct`

```go
func Setup(t *testing.T, req *Request) error {
	t.Log("entering listen-port test mode")
	return nil
}
```
