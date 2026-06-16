## Steps

- Run `http-proxy` with `--upstream-proxy http://127.0.0.1:19999` (nothing listening there) and NO `--fallback-direct`
- The process starts, logs a warning, then keeps listening
- Capture initial output and kill

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--upstream-proxy", "http://127.0.0.1:19987",
		"--listen-port", "19988",
	}
	return nil
}
```
