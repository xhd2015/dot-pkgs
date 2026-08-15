# Scenario

**Feature**: empty porcelain yields zero counts

```
"" -> ParsePorcelain -> all counts zero
```

## Steps

1. Set `req.Porcelain` to empty string.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Porcelain = ""
	return nil
}
```
