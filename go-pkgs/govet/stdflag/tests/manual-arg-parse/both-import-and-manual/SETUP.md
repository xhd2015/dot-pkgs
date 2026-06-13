## Preconditions
- The leaf directory contains `fixture.go` that both imports `"flag"` AND has manual flag parsing.

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
