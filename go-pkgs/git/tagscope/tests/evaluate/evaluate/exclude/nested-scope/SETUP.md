# Scenario

**Feature**: nested scopes both receive decisions when each has owned changes

```
sub/ and sub/nested/ both diff -> independent NextTag per scope
```

## Steps

1. Set tag names for `sub/` and `sub/nested/` scopes.
2. Inject owned tree diffs for both scopes.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"sub/v0.2.3", "sub/nested/v0.1.1"}
	req.HeadCommit = "head8888"
	req.ReleaseCommits = map[tagscope.TagScopeKey]string{
		tagscope.TagScopeKey("sub/"):         "release9999",
		tagscope.TagScopeKey("sub/nested/"): "releaseaaaa",
	}
	req.OwnedTrees = map[tagscope.TagScopeKey]tagscope.OwnedTreePair{
		tagscope.TagScopeKey("sub/"): ownedPair(
			tagscope.OwnedTree{"sub/pkg.go": "100644 aaa"},
			tagscope.OwnedTree{"sub/pkg.go": "100644 bbb"},
		),
		tagscope.TagScopeKey("sub/nested/"): ownedPair(
			tagscope.OwnedTree{"sub/nested/mod.go": "100644 ccc"},
			tagscope.OwnedTree{"sub/nested/mod.go": "100644 ddd"},
		),
	}
	return nil
}
```