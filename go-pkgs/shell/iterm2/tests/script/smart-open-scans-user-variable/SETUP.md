# Scenario

**Feature**: BuildScript scan matches path or user.koolTargetDir

```
BuildScript -> scan: path == targetDir OR user.koolTargetDir == targetDir
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-smart-scan-user")
	return nil
}
```