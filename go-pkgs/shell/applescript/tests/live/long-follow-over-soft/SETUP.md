# Scenario

**Live e2e**: long FollowUp (printf of shell-quoted body, SoftExceeded) via write text
often fails exact match — demonstrates the limitation.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Op = "live-long"
	return nil
}
```
