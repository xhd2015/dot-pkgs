# Scenario

**Feature**: redirect chain ends on universal `iTerm2-3_6_11.zip`

```
GET /stable/latest -> 302 /downloads/iTerm2-3_6_11.zip
  -> url ends with iTerm2-3_6_11.zip, version "3.6.11", no arm64/amd64
```

## Steps

1. Set `HTTPMode=redirect-zip` and fixed zip name.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.HTTPMode = "redirect-zip"
	req.FinalZipName = "iTerm2-3_6_11.zip"
	return nil
}
```
