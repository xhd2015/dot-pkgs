# Scenario

**Feature**: numeric release sorts above prereleases at the same version

```
[v0.0.1-beta, v0.0.1, v0.0.1-alpha] -> Tags v0.0.1, v0.0.1-beta, v0.0.1-alpha
```

## Steps

1. Set `req.Names` with same version release and prereleases.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Names = []string{"v0.0.1-beta", "v0.0.1", "v0.0.1-alpha"}
	return nil
}
```