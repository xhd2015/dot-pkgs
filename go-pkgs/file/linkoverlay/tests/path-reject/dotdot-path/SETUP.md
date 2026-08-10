# Scenario

**Feature**: `File.Path` containing `..` is rejected

```
Apply Files Path="foo/../../etc/passwd" -> error mentioning invalid / ..
```

## Steps

1. Single Files-only layer with path `foo/../../outside.txt`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	req.Layers = []LayerSpec{
		{
			Files: []FileSpec{
				{Path: "foo/../../outside.txt", Content: "nope"},
			},
		},
	}
	return nil
}
```
