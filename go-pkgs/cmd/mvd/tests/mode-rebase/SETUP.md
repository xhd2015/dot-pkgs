## Steps
- Log the rebase mode.

## Steps
- All tests in this mode exercise `mvd --rebase DIR NEW-DIR` to change the root key of a tracked entry.
- The new base directory becomes the first entry in the history chain, with the original entries preserved after it.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    t.Logf("mode: rebase")
    return nil
}
```
