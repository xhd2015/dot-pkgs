## Preconditions
- The leaf directory contains `fixture.go` with a switch inside a for loop, but no cases with `--` or `-` prefix.

## Steps
1. Read `fixture.go` from the current directory.
2. Set `req.Src` to the file contents.

```go
import (
	"os"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	data, err := os.ReadFile("fixture.go")
	if err != nil {
		return err
	}
	req.Src = string(data)
	return nil
}
```
