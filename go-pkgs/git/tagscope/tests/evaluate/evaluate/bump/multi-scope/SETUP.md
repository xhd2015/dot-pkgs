# Scenario

**Feature**: multiple scopes bump independently when each has owned changes

```
root v0.0.2 + sub/v0.2.3 both diff -> v0.0.3 and sub/v0.2.4
```

## Steps

1. Set tag names spanning root and `sub/` scopes.
2. Inject owned tree diffs for both scopes.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.2", "sub/v0.2.3"}
	req.HeadCommit = "head6666"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey(""):   "release5555",
		tagscope.TagScopeKey("sub/"): "release6666",
	}
	req.OwnedTrees = map[tagscope.TagScopeKey]tagscope.OwnedTreePair{
		tagscope.TagScopeKey(""): ownedPair(
			tagscope.OwnedTree{"README": "100644 aaa"},
			tagscope.OwnedTree{"README": "100644 bbb"},
		),
		tagscope.TagScopeKey("sub/"): ownedPair(
			tagscope.OwnedTree{"sub/lib.go": "100644 ccc"},
			tagscope.OwnedTree{"sub/lib.go": "100644 ddd"},
		),
	}
	return nil
}
```