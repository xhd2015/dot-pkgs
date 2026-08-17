# Scenario

**Feature**: go.mod without a `go` directive is an error

```
# module line only
go.mod without go line -> ModuleGoLine -> error
```

## Steps

1. Write `go.mod` with a module path and no `go` line.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	writeGoMod(t, req.ModDir, "module example.com/mod\n")
	return nil
}
```
