## Steps
- Log the list mode.

## Steps
- All tests in this mode exercise `mvd --list [SRC]` to display movement history.
- Without a SRC argument, all tracked projects are listed.
- With a SRC argument, only that specific project's history is shown, including markers for "(original)" and the current location.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Logf("mode: list")
	return nil
}
```
