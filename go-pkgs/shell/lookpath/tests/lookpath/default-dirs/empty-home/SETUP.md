# Scenario

**Feature**: empty home omits home-relative directories

```
DefaultDirs("") -> system bins only (/opt/homebrew/bin, /usr/local/bin)
  no "$HOME/..." style entries
```

## Steps

1. Set `DefaultDirsHome=""`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.DefaultDirsHome = ""
	return nil
}
```
