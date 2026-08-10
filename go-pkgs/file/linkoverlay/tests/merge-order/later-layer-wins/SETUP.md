# Scenario

**Feature**: two Dir layers, same top-level leaf — later layer wins; earlier base unchanged

```
base0/marker="early" then base1/marker="late"
  -> Apply(target, Dir base0, Dir base1)
  -> target/marker content "late"; base0/marker still "early"
```

## Steps

1. Materialize base0 with `marker=early` and base1 with `marker=late`.
2. Apply both as Dir-only layers in that order.

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
				{Path: "marker", Content: "early"},
			},
		},
		{
			DirRel: "base1",
			BaseFiles: []FileSpec{
				{Path: "marker", Content: "late"},
			},
		},
	}
	materializeLayers(t, req)
	return nil
}
```
