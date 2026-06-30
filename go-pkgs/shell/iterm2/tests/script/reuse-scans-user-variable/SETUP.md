# Scenario

**Feature**: reuse scan matches `path` or `user.koolTargetDir`

```
caller dir -> BuildReuseCurrentSessionScript -> scan: path == targetDir OR user.koolTargetDir == targetDir
```

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-scan-user")
	req.Mode = "reuse"
	return nil
}
```