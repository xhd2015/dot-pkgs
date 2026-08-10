# Scenario

**Feature**: multi-name order and Name fields match input order (no sort/dedupe)

```
Names=["beta", "alpha"] LookPathHits both
  -> items[0].Name=beta, items[1].Name=alpha (input order)
```

## Steps

1. Set two distinct names in non-alphabetical input order.
2. Inject LookPath hits for both under WorkDir.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"beta", "alpha"}
	beta := filepath.Join(req.WorkDir, "bin", "beta")
	alpha := filepath.Join(req.WorkDir, "bin", "alpha")
	writeExecutable(t, beta)
	writeExecutable(t, alpha)
	req.LookPathHits = map[string]string{
		"beta":  beta,
		"alpha": alpha,
	}
	return nil
}
```
