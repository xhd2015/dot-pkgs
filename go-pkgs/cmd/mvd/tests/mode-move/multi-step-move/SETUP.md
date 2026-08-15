# Scenario

Two sequential plain moves forming a chain.

mvd src d1 → [(src), (d1/src)]
mvd d1/src d2 → [(src), (d1/src), (d2/src)]

## Steps
- Create src, d1, and d2 directories.
- Perform the first move (src -> d1) by calling Run directly.
- Set req.Args for the second move (d1/src -> d2).

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	src := filepath.Join(req.WorkRoot, "src")
	d1 := filepath.Join(req.WorkRoot, "d1")
	d2 := filepath.Join(req.WorkRoot, "d2")
	mkdirAll(t, src)
	mkdirAll(t, d1)
	mkdirAll(t, d2)

	req.Args = []string{src, d1}
	resp, err := runMvd(t, d, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		t.Fatalf("first move failed: %s", resp.Output)
	}

	p1 := filepath.Join(d1, "src")
	req.Args = []string{p1, d2}
	return nil
}
```
