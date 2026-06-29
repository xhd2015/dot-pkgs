# Scenario

**Feature**: escape double quotes in follow-up commands

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "escape-command"
	req.EscapeInput = `echo "hi"`
	return nil
}
```