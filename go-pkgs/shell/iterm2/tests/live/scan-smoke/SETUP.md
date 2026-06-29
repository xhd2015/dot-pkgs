# Scenario

**Feature**: live / scan-smoke

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "live-smoke"
	return nil
}
```