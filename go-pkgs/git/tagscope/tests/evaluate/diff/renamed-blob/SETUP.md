# Scenario

**Feature**: same path with different blob oid is a change

```
path blob identity differs -> DiffOwnedTrees -> true
```

## Steps

1. Set same path key with different blob ids in old and new trees.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.OldTree = tagscope.OwnedTree{"README": "100644 aaa"}
	req.NewTree = tagscope.OwnedTree{"README": "100644 bbb"}
	return nil
}
```