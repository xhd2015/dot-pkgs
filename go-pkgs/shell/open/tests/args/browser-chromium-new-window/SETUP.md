# Scenario

**Feature**: a Chromium browser opens the URL in a new window, not a reused tab

```
BrowserArgs(url, "opera", true, "darwin")
  -> ["open", "-na", "Opera", "--args", "--new-window", url]
```

Opera is the regression this pins. It is a Chromium browser, but it used to be
missing from both the alias table and the Chromium-name check, so it fell through
to `open -na <app> <url>`: the running instance opened a tab in its existing
window and macOS switched Spaces to reach it. `-n` is what makes `--args` reach
the browser at all.

## Steps

1. Set `Kind=browser`, `GOOS=darwin`, a URL, the alias `opera`, and
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
	req.Browser = "opera"
	req.NewWindow = true
	return nil
}
```
