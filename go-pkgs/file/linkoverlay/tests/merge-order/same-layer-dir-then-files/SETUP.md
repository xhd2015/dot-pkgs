# Scenario

**Feature**: within one layer, Dir is applied first then Files — Files win on conflict

```
Layer{Dir: base with marker=A, Files: marker=B} -> Apply
  -> target/marker content B (real file, not write-through)
```

## Steps

1. Create base dir with `marker` content `A`.
2. Same layer Files overlay sets `marker` content `B`.

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
			DirRel: "base",
			BaseFiles: []FileSpec{
				{Path: "marker", Content: "A"},
			},
			Files: []FileSpec{
				{Path: "marker", Content: "B"},
			},
		},
	}
	materializeLayers(t, req)
	return nil
}
```
