# Scenario

**Feature**: RecognizeFile outcomes (success vs validation errors)

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "recognize_file"
	return nil
}
```
