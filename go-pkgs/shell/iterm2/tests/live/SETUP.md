# Scenario

**Feature**: live osascript compatibility smoke against iTerm2

```
BuildPathScanSmokeScript -> osascript -> session path probe -> "ok"
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "live-smoke"
	return nil
}
```