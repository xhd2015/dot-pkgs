# Scenario

**Feature**: empty tag list yields empty inventory

```
[] -> CollectFromNames -> empty CollectedTags
```

## Steps

1. Set `req.Names` to an empty slice.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{}
	return nil
}
```