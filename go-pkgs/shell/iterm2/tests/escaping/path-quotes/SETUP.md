# Scenario

**Feature**: escape double quotes in paths

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "escape-path"
	req.EscapeInput = `/tmp/"proj"`
	return nil
}
```