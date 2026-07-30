# Scenario

**Feature**: multi-line `ps -ax -o pid=,ppid=,command=` sample parses to rows

```
ps.sample.txt -> ParsePSOutput -> 4 rows with PID/PPID/Cmd
```

## Steps

1. Read `testdata/ps.sample.txt` into `req.PSOutput`.

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	path := filepath.Join(d.DOCTEST_CASE, "testdata", "ps.sample.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	req.PSOutput = data
	return nil
}
```
