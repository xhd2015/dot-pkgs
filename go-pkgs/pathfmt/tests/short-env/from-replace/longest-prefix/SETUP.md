# Scenario

**Feature**: longest matching env path prefix wins over a shorter parent alias

```
# env
X=/tmp/root
AI=/tmp/root/ai-workspace
path=/tmp/root/ai-workspace/src/foo
-> $AI/src/foo  (not $X/ai-workspace/src/foo)
```

## Steps

1. Create nested temp dirs: root and `ai-workspace` child.
2. Env: `X=root`, `AI=root/ai-workspace`.
3. Path: `root/ai-workspace/src/foo` (need not exist beyond Abs).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	root := t.TempDir()
	ai := filepath.Join(root, "ai-workspace")
	if err := os.MkdirAll(ai, 0o755); err != nil {
		return err
	}
	req.Env = []string{
		envPair("X", root),
		envPair("AI", ai),
	}
	req.Path = filepath.Join(ai, "src", "foo")
	return nil
}
```
