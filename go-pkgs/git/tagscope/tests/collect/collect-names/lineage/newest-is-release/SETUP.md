# Scenario

**Feature**: numeric release head equals latest release

```
[v0.0.1, v0.0.2] -> lineage Newest=LatestRelease=v0.0.2, no prerelease head
```

## Steps

1. Set `req.Names` with two ascending numeric releases.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.1", "v0.0.2"}
	return nil
}
```