# Scenario

**Feature**: fully missing nested path under cwd still shortens

```
# no .codex tree yet — only project root exists
path -> Short -> ".codex/hooks/agent-sessions-stop.sh"
```

## Steps

1. Set `req.Path` to a nested path whose intermediate dirs and leaf file are all missing.

```go
import (
	"path/filepath"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Path = filepath.Join(req.Path, ".codex", "hooks", "agent-sessions-stop.sh")
	return nil
}
```