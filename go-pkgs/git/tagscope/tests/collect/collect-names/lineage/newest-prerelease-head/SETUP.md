# Scenario

**Feature**: prerelease head with older numeric latest release

```
[v0.0.2, v0.0.3-alpha] -> lineage Newest=alpha, LatestRelease=v0.0.2
```

## Steps

1. Set `req.Names` with numeric release and newer prerelease.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.2", "v0.0.3-alpha"}
	return nil
}
```