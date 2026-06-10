## Steps
- Log the add mode.

## Steps
- All tests in this mode exercise `mvd --add DIR` to register a directory in the movement history without actually moving it.
- Adding the same directory twice is idempotent (the second add reports "already recorded").
- Adding a non-existent path fails with an error.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Logf("mode: add")
	return nil
}
```
