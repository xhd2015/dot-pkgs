# Scenario

**Feature**: a valid path runs the default handler once

```
PathConfig(path, cfg) -> Run(["open", path]) -> Result
```

## Steps

1. Set `Kind=path`, `GOOS=darwin`, and the fixture `Target`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Kind = "path"
	req.GOOS = "darwin"
	return nil
}
```
