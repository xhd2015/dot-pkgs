# Scenario

**Feature**: full InstallLatest pipeline with fake HTTP zip under Home Applications

```
fake /stable/latest -> zip iTerm2-3_6_11.zip
  -> Result.AppPath = Home/Applications/iTerm.app
  -> Version 3.6.11; Register called; VerifyInstalled passes
```

## Steps

1. Set `HTTPMode=install-latest`, leave TargetApp empty for default home path.
2. Root Run forces `SkipScriptable` and injects Register + HTTPClient + LatestURL.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.HTTPMode = "install-latest"
	req.FinalZipName = "iTerm2-3_6_11.zip"
	req.TargetApp = "" // default Home/Applications/iTerm.app
	req.SkipScriptable = true
	return nil
}
```
