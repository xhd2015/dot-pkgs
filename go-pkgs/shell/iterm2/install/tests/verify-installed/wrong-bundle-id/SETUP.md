# Scenario

**Feature**: wrong CFBundleIdentifier → VerifyInstalled error

```
Info.plist CFBundleIdentifier=com.example.wrong -> error
```

## Steps

1. Override BundleID to a non-iTerm value.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.BundleIDOverride = "com.example.wrong"
	req.OmitMacOSBinary = false
	return nil
}
```
