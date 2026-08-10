# Scenario

**Feature**: absolute `File.Path` is rejected

```
Apply Files Path="/tmp/abs-overlay.txt" -> error mentioning absolute / invalid
```

## Steps

1. Single Files-only layer with an absolute path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	abs := filepath.Join(string(filepath.Separator), "tmp", "linkoverlay-abs-reject.txt")
	req.Layers = []LayerSpec{
		{
			Files: []FileSpec{
				{Path: abs, Content: "nope"},
			},
		},
	}
	return nil
}
```
