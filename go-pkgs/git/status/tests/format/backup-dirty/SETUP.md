# Scenario

**Feature**: non-zero counts format as backup dirty string

```
Modified=2, Untracked=1 -> Format(FormatBackup) -> dirty (2 modified, 1 untracked)
```

## Steps

1. Set counts matching machine backup style.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, req *Request) error {
	req.Counts = status.Counts{Modified: 2, Untracked: 1}
	return nil
}
```
