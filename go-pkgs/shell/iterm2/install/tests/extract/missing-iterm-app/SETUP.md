# Scenario

**Feature**: zip without `iTerm.app` → error

```
zip(not-iterm/readme.txt) -> ExtractApp -> error
```

## Steps

1. Set `ZipWithoutApp=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.ZipWithoutApp = true
	return nil
}
```
