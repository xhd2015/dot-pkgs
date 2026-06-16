## Steps

- Run `http-proxy` with `--upstream-proxy http://127.0.0.1:19999` (nothing listening) and `--fallback-direct`
- The process starts, logs fallback, then we capture output and kill

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{
		"--upstream-proxy", "http://127.0.0.1:19985",
		"--fallback-direct",
		"--listen-port", "19986",
	}
	return nil
}
```
