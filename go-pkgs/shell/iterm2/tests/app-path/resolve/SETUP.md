# Scenario

**Feature**: ResolveAppPathWith candidate order + env strictness

```
inject Getenv / Home / IsApp -> ResolveAppPathWith -> path or ""
```

## Steps

1. Phase `resolve-app`.
2. Leaves set EnvSet/EnvValue, HomeDir, ExistingDirs.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "resolve-app"
	return nil
}
```
