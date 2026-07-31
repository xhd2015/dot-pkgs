# Scenario

**Feature**: parse multi-line session list dump with tab indices and TTY fields

```
fixture sessions.txt (TSV lines)
  -> ParseSessionListOutput
  -> []SessionRef with WindowID, TabIndex, TTY, Name, SessionID
```

## Steps

1. Phase `parse-session-list`.
2. Load `ListOutput` from `sessions.txt` beside this leaf (case dir).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "parse-session-list"
	path := filepath.Join(d.DOCTEST_CASE, "sessions.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	req.ListOutput = string(data)
	return nil
}
```
