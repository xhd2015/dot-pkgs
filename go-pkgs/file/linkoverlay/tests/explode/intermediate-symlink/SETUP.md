# Scenario

**Feature**: explode `.config` seed so sibling `other/x` stays visible after writing `tool/c`

```
Layer0 Dir: .config/other/x = O
Layer1 Files: .config/tool/c = C
  -> target/.config is real dir (exploded)
  -> target/.config/other/x readable O (re-linked)
  -> target/.config/tool/c content C
```

## Steps

1. Materialize base0 with only `.config/other/x`.
2. Apply Files overlay at `.config/tool/c` (forces explode of intermediate `.config`).

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Layers = []LayerSpec{
		{
			DirRel: "base0",
			BaseFiles: []FileSpec{
				{Path: ".config/other/x", Content: "O"},
			},
		},
		{
			Files: []FileSpec{
				{Path: ".config/tool/c", Content: "C"},
			},
		},
	}
	materializeLayers(t, req)
	return nil
}
```
