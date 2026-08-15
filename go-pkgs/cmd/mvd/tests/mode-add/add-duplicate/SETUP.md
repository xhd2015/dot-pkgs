# Scenario

Adding the same directory twice is idempotent.

mvd --add tracked → [(tracked)]
mvd --add tracked → [(tracked)]  (no change)

## Steps
- Create a directory and add it with --add.
- Add the same directory again with --add.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	dir := filepath.Join(req.WorkRoot, "tracked")
	mkdirAll(t, dir)
	req.Args = []string{"--add", dir}
	resp, err := runMvd(t, d, req)
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
