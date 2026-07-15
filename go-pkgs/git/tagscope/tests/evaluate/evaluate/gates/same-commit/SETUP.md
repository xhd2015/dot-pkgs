# Scenario

**Feature**: release tag commit equals HEAD commit

```
LatestRelease commit == HeadCommit -> gate same-commit (before diff)
```

## Steps

1. Set tag names with numeric release baseline.
2. Set `HeadCommit` equal to release commit for root scope.
3. Inject owned trees that differ (gate still fires before diff).

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"v0.0.2"}
	req.HeadCommit = "commit-same"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey(""): "commit-same",
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