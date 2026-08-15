# Scenario

**Feature**: reuse scan matches `path` or `user.koolTargetDir`

```
caller dir -> BuildReuseCurrentSessionScript -> scan: path == targetDir OR user.koolTargetDir == targetDir
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-scan-user")
	req.Mode = "reuse"
	return nil
}
```