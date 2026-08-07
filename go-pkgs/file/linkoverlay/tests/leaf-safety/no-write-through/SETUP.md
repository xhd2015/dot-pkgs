# Scenario

**Feature**: replacing a seeded leaf symlink writes a new file; base content stays original

```
Dir base: secret=ORIGINAL
Files: secret=OVERLAY
  -> target/secret = OVERLAY (regular)
  -> base/secret still ORIGINAL
```

## Steps

1. Materialize base with `secret=ORIGINAL`.
2. Overlay Files `secret=OVERLAY` in a later layer (or same layer Files after Dir).

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
				{Path: "secret", Content: "ORIGINAL"},
			},
		},
		{
			Files: []FileSpec{
				{Path: "secret", Content: "OVERLAY"},
			},
		},
	}
	materializeLayers(t, req)
	return nil
}
```
