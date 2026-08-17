# Scenario

**Feature**: `go 1.19` in go.mod becomes `go1.19`

```
# major.minor go directive
go.mod "go 1.19" -> ModuleGoLine -> go1.19
```

## Steps

1. Write `go.mod` with `go 1.19`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	writeGoMod(t, req.ModDir, "module example.com/mod\n\ngo 1.19\n")
	return nil
}
```
