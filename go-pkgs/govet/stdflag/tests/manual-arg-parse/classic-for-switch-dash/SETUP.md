## Preconditions
- The leaf directory contains `fixture.go` with a classic `for i := 0; i < len(args); i++` loop containing a `switch` with `--` prefixed cases.

## Steps
1. Read `fixture.go` from the current directory.
2. Set `req.Src` to the file contents.

```go
import (
	"os"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	data, err := os.ReadFile("fixture.go")
	if err != nil {
		return err
	}
	req.Src = string(data)
	return nil
}
```
