# Scenario

**Feature**: reuse miss branch registers `user.koolTargetDir` for back-to-back reuse

```
caller dir -> BuildReuseCurrentSessionScript -> miss branch: cd + set user.koolTargetDir
```

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Dir = absDir(t, "/tmp/iterm2-reuse-register")
	req.Mode = "reuse"
	return nil
}
```