## Steps
- Create a directory and add it with --add.
- Add the same directory again with --add.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "tracked")
	mkdirAll(t, dir)
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add: %s", resp.Output)
	}

	req.Args = []string{"--add", dir}
	return nil
}
```
