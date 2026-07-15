# Scenario

**Feature**: missing path in AtHead is a change

```
path present only in old tree -> DiffOwnedTrees -> true
```

## Steps

1. Set `OldTree` with one path and `NewTree` empty.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, req *Request) error {
	req.OldTree = tagscope.OwnedTree{"README": "100644 aaa"}
	req.NewTree = tagscope.OwnedTree{}
	return nil
}
```