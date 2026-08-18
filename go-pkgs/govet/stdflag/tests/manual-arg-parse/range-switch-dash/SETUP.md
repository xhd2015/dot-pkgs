## Preconditions
- The leaf directory contains `fixture.go` with a `for _, arg := range args` loop containing a `switch` with `--` prefixed case.

## Steps
1. Read `fixture.go` from the leaf directory via `fixtureFile` (`DOCTEST_CASE`).
2. Set `req.Src` to the file contents.

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	data, err := os.ReadFile(fixtureFile(d, "fixture.go"))
	if err != nil {
		return err
	}
	req.Src = string(data)
	return nil
}
```
