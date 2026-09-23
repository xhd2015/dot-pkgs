# Scenario

**Feature**: no `--browser` still opens a new window in the default browser

```
BrowserConfig(url, "", true, cfg) with DefaultBrowser -> "Opera"
  -> ["open", "-na", "Opera", "--args", "--new-window", url]
```

Without this, a bare `open` fell to `{"open", "-n", url}`: `-n` only starts a new
instance, and the URL is still delivered to the running browser as a document, so
it opens a tab in that window and macOS switches Spaces to show it. Resolving the
default browser and asking *it* for a new window is what makes the flagless form
behave like `--browser=<default>`.

## Steps

1. Set `Operation=run`, `Kind=browser`, `GOOS=darwin`, a URL,
   `NewWindow=true`, an empty `Browser`, and a resolvable `DefaultBrowser`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Operation = "run"
	req.Kind = "browser"
	req.GOOS = "darwin"
	req.URL = "http://127.0.0.1:8422/projects/eluc"
	req.Browser = ""
	req.NewWindow = true
	req.DefaultBrowser = "Opera"
	return nil
}
```
