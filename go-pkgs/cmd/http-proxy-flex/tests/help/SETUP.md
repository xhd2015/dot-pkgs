## Preconditions

- The `http-proxy` binary can be compiled

## Steps

- Testing the `--help` flag behavior

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--help"}
	return nil
}
```
