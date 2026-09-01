# Scenario

**Feature**: dirty main blocks remote sync

```
dirty main + origin -> MergeBack errors with main-sync
```

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := os.WriteFile(filepath.Join(req.MainRepo, "dirty-main.txt"), []byte("x\n"), 0644); err != nil {
		return err
	}
	return nil
}
```
