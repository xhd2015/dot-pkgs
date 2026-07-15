# Scenario

**Feature**: sub-scope owned file changed at HEAD

```
sub/v0.2.3 baseline + sub/lib.go diff -> NextTag=sub/v0.2.4
```

## Steps

1. Set `sub/` scope tag `sub/v0.2.3`.
2. Inject owned trees with change under `sub/` prefix only.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, req *Request) error {
	req.Names = []string{"sub/v0.2.3"}
	req.HeadCommit = "head5555"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey("sub/"): "release4444",
	}
	req.OwnedTrees = map[tagscope.TagScopeKey]tagscope.OwnedTreePair{
		tagscope.TagScopeKey("sub/"): ownedPair(
			tagscope.OwnedTree{"sub/lib.go": "100644 aaa"},
			tagscope.OwnedTree{"sub/lib.go": "100644 bbb"},
		),
	}
	return nil
}
```