# Scenario

**Feature**: osascript failure wrapped and returned

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = t.TempDir()
	req.OsascriptFail = true
	return nil
}
```