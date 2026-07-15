# Scenario

**Feature**: prerelease-only scope has nil latest release

```
[v0.0.1-alpha] -> lineage LatestRelease=nil, HasPrereleaseHead=true
```

## Steps

1. Set `req.Names` to a single prerelease tag.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.1-alpha"}
	return nil
}
```