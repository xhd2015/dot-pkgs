# Scenario

**Feature**: termenv OSC 11 (ST) + CSI 6n pair is removed from surrounding text

```
# termenv pair
Data = notice + ESC]11;?ST + ESC[6n + warning
  -> display noticewarning
```

## Steps

1. Set `req.Data` to the termenv probe pair wrapped in text.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Data = []byte("notice\x1b]11;?\x1b\\\x1b[6nwarning")
	return nil
}
```
