## Steps
- Put HOME and X on the mvd child env (`req.ExtraEnv`) for lls dollar-expansion.
- HOME points to `req.WorkRoot/.lls-home` where the lls config with `"envs":["X"]` resides.
- X resolves to `req.WorkRoot/projects`.
- Do not use `t.Setenv`: doctest leaves always call `t.Parallel()`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	homeDir := filepath.Join(req.WorkRoot, ".lls-home")
	req.ExtraEnv = []string{
		"HOME=" + homeDir,
		"X=" + filepath.Join(req.WorkRoot, "projects"),
	}
	return nil
}
```
