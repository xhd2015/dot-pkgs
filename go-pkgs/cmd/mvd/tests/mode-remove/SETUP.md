## Steps

## Steps
- All tests in this mode exercise `mvd --rm | --remove` to delete a tracked entry from history.
- Without `--force` (`-f`), removing an entry that has movement history (more than one location) is rejected.
- With `--force`, the entry and all its history is cleared.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    t.Logf("mode: remove")
    return nil
}
```
