## Steps
- Create src, d1, and d2 directories.
- First move src to d1 to register it in history.
- Then use the original root path to move from d1 to d2.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	d1 := filepath.Join(req.WorkRoot, "d1")
	d2 := filepath.Join(req.WorkRoot, "d2")
	mkdirAll(t, src)
	mkdirAll(t, d1)
	mkdirAll(t, d2)

	req.Args = []string{src, d1}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("first move: %s", resp.Output)
	}

	req.Args = []string{src, d2}
	return nil
}
```
