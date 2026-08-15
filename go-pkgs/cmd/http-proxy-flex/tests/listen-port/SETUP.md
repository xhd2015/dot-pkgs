## Preconditions

- The `http-proxy` binary can be compiled
- No process is listening on port 19999 (unused by other tests)

## Steps

- Testing the `--listen-port` flag: default vs custom values
- Both children use `--upstream-proxy` pointing to a dead port (default flex)

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Log("entering listen-port test mode")
	return nil
}
```
