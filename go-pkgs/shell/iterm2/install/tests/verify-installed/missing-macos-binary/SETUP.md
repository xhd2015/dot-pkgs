# Scenario

**Feature**: missing MacOS binary → VerifyInstalled error

```
iTerm.app with Info.plist but no Contents/MacOS/iTerm2 -> error
```

## Steps

1. Set `OmitMacOSBinary=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.OmitMacOSBinary = true
	return nil
}
```
