# Scenario

**Feature**: unset KEY present in base → KEY absent (S4)

```
# S4 unset present
Base has SECRET=x and PATH=/bin
Unset = SECRET
  -> MergeProcessEnv
  -> SECRET absent; PATH preserved
```

## Steps

1. Base: `SECRET=s3cr3t`, `PATH=/bin`, `HOME=/tmp/home`.
2. Unset: `SECRET`.
3. Set empty.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Base = []string{
		"SECRET=s3cr3t",
		"PATH=/bin",
		"HOME=/tmp/home",
	}
	req.Set = nil
	req.Unset = []string{"SECRET"}
	return nil
}
```
