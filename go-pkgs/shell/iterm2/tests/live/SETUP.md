# Scenario

**Feature**: live osascript compatibility smoke against iTerm2

```
BuildPathScanSmokeScript -> osascript -> session path probe -> "ok"
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "live-smoke"
	return nil
}
```