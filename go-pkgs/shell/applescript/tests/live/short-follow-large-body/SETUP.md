# Scenario

**Live e2e**: short FollowUp (`bash script.sh`) + multi-KB Chinese body on disk → exact match.

Demonstrates control: write text stays short; body unlimited via file.

## Steps

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Op = "live-short"
	return nil
}
```
