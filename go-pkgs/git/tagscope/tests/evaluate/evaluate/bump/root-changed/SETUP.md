# Scenario

**Feature**: root scope owned file changed at HEAD

```
v0.0.2 baseline + README blob diff -> NextTag=v0.0.3
```

## Steps

1. Set root scope tag `v0.0.2`.
2. Inject owned trees with README blob change at root scope.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.2"}
	req.HeadCommit = "head4444"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey(""): "release3333",
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