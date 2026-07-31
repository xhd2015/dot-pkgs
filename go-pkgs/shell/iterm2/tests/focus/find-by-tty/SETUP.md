# Scenario

**Feature**: pure FindByTTY filter over SessionRef list

```
[]SessionRef + query TTYs
  -> FindByTTY (NormalizeTTY on both sides)
  -> matching refs (union; stable order of refs)
```

## Steps

1. Leaves set Phase `find-by-tty`, fixture `Refs`, and `QueryTTYs`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Phase = "find-by-tty"
	return nil
}
```
