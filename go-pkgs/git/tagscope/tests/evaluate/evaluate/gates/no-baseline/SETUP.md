# Scenario

**Feature**: scope with only prerelease tags has no baseline release

```
[v0.0.1-alpha] -> LatestRelease=nil -> gate no-baseline
```

## Steps

1. Set tag names to prerelease-only root scope.
2. Inject owned trees with a diff (gate fires before diff).

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.1-alpha"}
	req.HeadCommit = "head1111"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey(""): "release0000",
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