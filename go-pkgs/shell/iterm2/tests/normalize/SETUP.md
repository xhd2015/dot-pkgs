# Scenario

**Feature**: existing paths normalize via EvalSymlinks before script build

```
symlink dir -> OpenConfig -> EvalSymlinks -> targetDir in AppleScript
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Phase = "open-config"
	return nil
}
```