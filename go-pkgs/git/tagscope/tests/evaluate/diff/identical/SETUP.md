# Scenario

**Feature**: identical owned trees report no change

```
same path->blob map -> DiffOwnedTrees -> false
```

## Steps

1. Set `OldTree` and `NewTree` to identical maps.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	tree := tagscope.OwnedTree{
		"README": "100644 aaa",
		"go.mod": "100644 bbb",
	}
	req.OldTree = tree
	req.NewTree = tree
	return nil
}
```