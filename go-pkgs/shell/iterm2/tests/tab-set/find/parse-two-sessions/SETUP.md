# Scenario

**Feature**: parse two sessions in the same set with different tab ids

```
fixture (two tab-separated lines)
  -> ParseTabSetFindOutput
  -> []TabSessionRef length 2; TabIDs t1,t2; SetName bots
```

## Steps

1. Phase `parse-find`.
2. Load `FindOutput` from `two-sessions.txt` fixture (or inline equivalent).

```go
import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Phase = "parse-find"
	// Process cwd is undetermined; join fixture to leaf case dir.
	path := filepath.Join(d.DOCTEST_CASE, "two-sessions.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	req.FindOutput = string(data)
	return nil
}
```
