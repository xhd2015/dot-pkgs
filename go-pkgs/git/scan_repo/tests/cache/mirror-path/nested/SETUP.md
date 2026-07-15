# Scenario

**Feature**: multi-segment absolute path yields nested mirror path segments

```
# nested real path -> nested dirs under mirror/
real:  /Users/xhd2015/Projects/org/team/repo
cache: <CacheRoot>/mirror/Users/xhd2015/Projects/org/team/repo/entry.json
```

## Steps

1. Set `req.RealPath` to a multi-segment absolute path
   `/Users/xhd2015/Projects/org/team/repo`.

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RealPath = "/Users/xhd2015/Projects/org/team/repo"
	return nil
}
```
