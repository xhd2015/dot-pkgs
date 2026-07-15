# Scenario

**Feature**: new path in AtHead is a change

```
path present only in new tree -> DiffOwnedTrees -> true
```

## Steps

1. Set `OldTree` empty and `NewTree` with one path.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/tagscope"
)

func Setup(t *testing.T, req *Request) error {
	req.OldTree = tagscope.OwnedTree{}
	req.NewTree = tagscope.OwnedTree{"README": "100644 aaa"}
	return nil
}
```