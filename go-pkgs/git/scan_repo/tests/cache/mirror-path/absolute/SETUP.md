# Scenario

**Feature**: absolute real path maps under mirror without a leading empty segment

```
# abs real path strips leading separator before joining under mirror/
real:  /Users/xhd2015/Projects/foo
cache: <CacheRoot>/mirror/Users/xhd2015/Projects/foo/entry.json
```

## Steps

1. Set `req.RealPath` to a fixed absolute Unix-style path
   `/Users/xhd2015/Projects/foo` (need not exist).

```go
import "testing"

func Setup(t *testing.T, req *Request) error {
	req.RealPath = "/Users/xhd2015/Projects/foo"
	return nil
}
```
