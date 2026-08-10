# Scenario

**Feature**: empty appPath → bare tell application "iTerm2"

```
AppPath=""
  -> TellApplicationHeader
  -> tell application "iTerm2"
```

## Steps

1. Leave AppPath empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.AppPath = ""
	return nil
}
```
