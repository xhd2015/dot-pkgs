## Steps
- Log the move mode.

## Steps
- All tests in this mode exercise the default `mvd SRC DST` command (no flags, two positional arguments).
- Resolution priority for SRC: unique root basename → alias → absolute path.
- If DST is an existing directory, the source is moved inside it (basename join).
- If DST does not exist, it becomes the new path directly.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Logf("mode: move")
	return nil
}
```
