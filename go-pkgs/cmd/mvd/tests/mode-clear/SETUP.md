## Steps
- Log the clear mode.

## Steps
- All tests in this mode exercise `mvd --clear SRC` to clear the movement history for a tracked entry.
- After clearing, the history file contains no record of the project.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
    t.Logf("mode: clear")
    return nil
}
```
