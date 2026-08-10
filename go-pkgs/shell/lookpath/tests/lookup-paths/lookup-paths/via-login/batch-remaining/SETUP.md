# Scenario

**Feature**: two remaining names both resolve via login; From set on each

```
Names=["tool-a", "tool-b"] all cheap stages miss
  -> RunLogin resolves both
  -> items found, From="bash" for each
```

## Steps

1. Two names; no PATH/ExtraDirs hits.
2. LoginStdoutByName maps bash → each name → distinct absolute path.

```go
import (
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Names = []string{"tool-a", "tool-b"}
	pathA := filepath.Join(req.WorkDir, "login", "bash", "tool-a")
	pathB := filepath.Join(req.WorkDir, "login", "bash", "tool-b")
	req.LoginStdoutByName = map[string]map[string]string{
		"bash": {
			"tool-a": pathA,
			"tool-b": pathB,
		},
	}
	return nil
}
```
