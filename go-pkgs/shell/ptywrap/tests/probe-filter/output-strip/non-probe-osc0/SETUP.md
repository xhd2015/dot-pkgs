# Scenario

**Feature**: OSC 0 window title is not treated as a color probe

```
# OSC 0
Data = ESC]0;title BEL + hello
  -> display unchanged
```

## Steps

1. Set `req.Data` to an OSC 0 sequence plus text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("\x1b]0;title\x07hello")
	return nil
}
```
