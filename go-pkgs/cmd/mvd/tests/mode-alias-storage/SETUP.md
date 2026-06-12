## Steps
- Log the alias storage mode.

## Description
- Tests verifying aliases are stored inside history.json instead of a separate aliases.json file.
- Aliases appear in the per-project `aliases` array within each ProjectEntry.
- After this change, mvd must never create aliases.json.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	t.Logf("mode: alias-storage")
	return nil
}
```
