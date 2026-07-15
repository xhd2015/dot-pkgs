# Scenario

**Feature**: prerelease head blocks release bump even when older numeric release exists

```
[v0.0.2, v0.0.3-rc1] -> HasPrereleaseHead -> gate prerelease-head
```

## Steps

1. Set tag names with numeric latest release and newer prerelease head.
2. Inject owned trees with changes (gate fires before diff).

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.2", "v0.0.3-rc1"}
	req.HeadCommit = "head2222"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey(""): "release1111",
	}
	req.OwnedTrees = map[tagscope.TagScopeKey]tagscope.OwnedTreePair{
		tagscope.TagScopeKey(""): ownedPair(
			tagscope.OwnedTree{"README": "100644 aaa"},
			tagscope.OwnedTree{"README": "100644 bbb"},
		),
	}
	return nil
}
```