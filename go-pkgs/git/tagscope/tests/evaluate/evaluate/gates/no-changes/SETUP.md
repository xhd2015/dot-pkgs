# Scenario

**Feature**: owned trees identical at release and HEAD

```
DiffOwnedTrees false -> gate no-changes
```

## Steps

1. Set tag names with numeric release baseline.
2. Set release and HEAD commits to different hashes.
3. Inject identical owned trees for root scope.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.2"}
	req.HeadCommit = "head3333"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey(""): "release2222",
	}
	tree := tagscope.OwnedTree{
		"README": "100644 aaa",
		"go.mod": "100644 ccc",
	}
	req.OwnedTrees = map[tagscope.TagScopeKey]tagscope.OwnedTreePair{
		tagscope.TagScopeKey(""): ownedPair(tree, tree),
	}
	return nil
}
```