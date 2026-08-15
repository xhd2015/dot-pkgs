## Steps

- Run `http-proxy` with `--upstream-proxy` pointing at a dead port and `--no-fallback-direct`
- The process starts, logs a warning, then keeps listening
- Capture initial output and kill

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{
		"--upstream-proxy", "http://127.0.0.1:19987",
		"--no-fallback-direct",
		"--listen-port", "19988",
	}
	return nil
}
```
