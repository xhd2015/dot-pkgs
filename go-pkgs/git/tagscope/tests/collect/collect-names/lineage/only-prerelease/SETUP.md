# Scenario

**Feature**: prerelease-only scope has nil latest release

```
[v0.0.1-alpha] -> lineage LatestRelease=nil, HasPrereleaseHead=true
```

## Steps

1. Set `req.Names` to a single prerelease tag.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.1-alpha"}
	return nil
}
```