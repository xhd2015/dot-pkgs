# Scenario

**Feature**: zero counts format as clean

```
empty Counts -> Format(FormatBackup) -> "clean"
```

## Steps

1. Leave `req.Counts` at zero value.

```go
import (
	"testing"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/status"
)

func Setup(t *testing.T, req *Request) error {
	req.Counts = status.Counts{}
	return nil
}
```
