# Scenario

**Feature**: Firefox gets its own new-window switch spelling

```
BrowserArgs(url, "firefox", true, "darwin")
  -> ["open", "-na", "Firefox", "--args", "-new-window", url]
```

Firefox takes a single-dash `-new-window`, unlike the Chromium family's
`--new-window`. It is a separate family rather than an unknown app so the flag is
not guessed.

## Steps

1. Set `Kind=browser`, `GOOS=darwin`, a URL, the alias `firefox`, and
   `NewWindow=true`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "browser"
	req.GOOS = "darwin"
	req.URL = "http://127.0.0.1:8422/projects/eluc"
	req.Browser = "firefox"
	req.NewWindow = true
	return nil
}
```
