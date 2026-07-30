# Scenario

**Feature**: `lsof -Fn` sample yields unique absolute name paths in first-seen order

```
lsof.fn.sample.txt -> ParseLsofFn -> unique absolute n-paths
```

## Steps

1. Read `testdata/lsof.fn.sample.txt` into `req.LsofOutput`.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	path := filepath.Join(d.DOCTEST_CASE, "testdata", "lsof.fn.sample.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	req.LsofOutput = data
	return nil
}
```
