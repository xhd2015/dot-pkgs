## Steps
- Create two project directories proj1 and proj2.
- Add both with --add.
- Request --list (no filter).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	s1 := filepath.Join(req.WorkRoot, "proj1")
	s2 := filepath.Join(req.WorkRoot, "proj2")
	mkdirAll(t, s1)
	mkdirAll(t, s2)

	req.Args = []string{"--add", s1}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add s1: %s", resp.Output)
	}

	req.Args = []string{"--add", s2}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("add s2: %s", resp.Output)
	}

	req.Args = []string{"--list"}
	return nil
}
```
