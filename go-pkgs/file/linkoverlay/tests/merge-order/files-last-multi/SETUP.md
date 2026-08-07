# Scenario

**Feature**: multi-layer Dir A + Dir B + Files pack — sparse Files beat dirs on conflict

```
Dir baseA (pack=A) + Dir baseB (pack=B, extra=B) + Files pack=C
  -> target/pack = C; target/extra follows baseB
```

## Steps

1. Layer0 Dir: `pack=content-A`.
2. Layer1 Dir: `pack=content-B`, `extra=from-B`.
3. Layer2 Files only: `pack=content-C`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Layers = []LayerSpec{
		{
			DirRel: "base-a",
			BaseFiles: []FileSpec{
				{Path: "pack", Content: "content-A"},
			},
		},
		{
			DirRel: "base-b",
			BaseFiles: []FileSpec{
				{Path: "pack", Content: "content-B"},
				{Path: "extra", Content: "from-B"},
			},
		},
		{
			// Files-only layer (no Dir)
			Files: []FileSpec{
				{Path: "pack", Content: "content-C"},
			},
		},
	}
	materializeLayers(t, req)
	return nil
}
```
