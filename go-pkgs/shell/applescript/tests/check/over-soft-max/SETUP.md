# Scenario

SoftMax+1 bytes → SoftExceeded.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "check"
	req.CheckExactLen = applescript.WriteTextSoftMaxBytes + 1
	return nil
}
```
