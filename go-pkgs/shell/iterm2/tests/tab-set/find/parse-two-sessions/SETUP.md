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
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	req.Phase = "parse-find"
	data, err := os.ReadFile("two-sessions.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	req.FindOutput = string(data)
	return nil
}
```
