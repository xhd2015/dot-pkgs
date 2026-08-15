## Steps
- Log the error mode.

## Steps
- All tests in this mode verify that mvd produces appropriate errors for invalid inputs.
- Moving a non-existent path or an unrecognized basename both result in errors.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Logf("mode: error")
	return nil
}
```
