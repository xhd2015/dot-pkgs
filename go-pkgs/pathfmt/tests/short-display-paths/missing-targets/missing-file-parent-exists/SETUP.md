# Scenario

**Feature**: missing leaf file with existing parent dirs shortens under cwd

```
# parent dirs exist, leaf file missing (integration hook script)
path -> Short -> ".codex/hooks/agent-sessions-stop.sh"
```

## Steps

1. Create `.codex/hooks/` under project root (no stop script file).
2. Set `req.Path` to the missing script absolute path.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	hooksDir := filepath.Join(req.Path, ".codex", "hooks")
	mkdirAll(t, hooksDir)
	req.Path = filepath.Join(hooksDir, "agent-sessions-stop.sh")
	return nil
}
```