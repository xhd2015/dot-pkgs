# Scenario

**Feature**: two Dir bases with disjoint top-level names (including a dot entry)

```
base-a: a.txt, .config/tool  +  base-b: b.txt
  -> ApplyDirs(target, base-a, base-b)
  -> target/a.txt, target/b.txt, target/.config as abs symlinks into respective bases
```

## Steps

1. Create base-a with `a.txt` and `.config/tool`.
2. Create base-b with `b.txt`.
3. Call ApplyDirs with both bases in order.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Layers = []LayerSpec{
		{
			DirRel: "base-a",
			BaseFiles: []FileSpec{
				{Path: "a.txt", Content: "from-a"},
				{Path: ".config/tool", Content: "cfg-a"},
			},
		},
		{
			DirRel: "base-b",
			BaseFiles: []FileSpec{
				{Path: "b.txt", Content: "from-b"},
			},
		},
	}
	req.DirsRel = []string{"base-a", "base-b"}
	materializeLayers(t, req)
	return nil
}
```
