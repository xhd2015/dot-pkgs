# Scenario

**Feature**: `OpenConfig` validates input and runs injectable osascript

```
caller dir -> OpenConfig(Config) -> stat dir -> BuildScript -> Osascript hook
```

## Steps

1. Set `req.Phase` to `open-config`.
2. Provide temp directory or file fixtures per leaf.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "open-config"
	return nil
}
```