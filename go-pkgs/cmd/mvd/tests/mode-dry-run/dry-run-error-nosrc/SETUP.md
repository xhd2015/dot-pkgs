# Scenario

--dry-run with non-existent src: validation error still fires.

mvd --dry-run no-such-dir dst → error → does not exist

## Steps
- Try to move a non-existent source with `--dry-run`.
- Validation should still fail — `--dry-run` does not suppress errors.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	nonexistent := filepath.Join(req.WorkRoot, "no-such-dir")
	dst := filepath.Join(req.WorkRoot, "dst")
	mkdirAll(t, dst)
	req.Args = []string{"--dry-run", nonexistent, dst}
	return nil
}
```
