# Scenario

**Feature**: change only under nested child scope does not affect parent owned trees

```
sub/ unchanged + sub/nested/ diff -> sub/ no-changes, nested bumps
```

## Steps

1. Set tag names for `sub/` and `sub/nested/` scopes.
2. Inject identical trees for `sub/` scope and diff for `sub/nested/` scope.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"sub/v0.2.3", "sub/nested/v0.1.1"}
	req.HeadCommit = "head7777"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey("sub/"):         "release7777",
		tagscope.TagScopeKey("sub/nested/"): "release8888",
	}
	subTree := tagscope.OwnedTree{"sub/pkg.go": "100644 aaa"}
	nestedRelease := tagscope.OwnedTree{"sub/nested/mod.go": "100644 bbb"}
	nestedHead := tagscope.OwnedTree{"sub/nested/mod.go": "100644 ccc"}
	req.OwnedTrees = map[tagscope.TagScopeKey]tagscope.OwnedTreePair{
		tagscope.TagScopeKey("sub/"):         ownedPair(subTree, subTree),
		tagscope.TagScopeKey("sub/nested/"): ownedPair(nestedRelease, nestedHead),
	}
	return nil
}
```