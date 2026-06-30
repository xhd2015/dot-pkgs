# Scenario

**Feature**: BuildScript scan matches path or user.koolTargetDir

```
BuildScript -> scan: path == targetDir OR user.koolTargetDir == targetDir
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-smart-scan-user")
	return nil
}
```