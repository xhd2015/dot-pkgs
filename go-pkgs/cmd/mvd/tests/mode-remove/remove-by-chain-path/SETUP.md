# Scenario

Remove a non-root path from a plain-move chain, preserving the root entry.

mvd --add repo → [{repo}]
mvd repo dst → [{repo}, {dst/repo}]
mvd --rm dst/repo → [{repo}]   (moved path removed, root preserved)

## Steps
- Create a repo directory at `repo` and add it to tracking.
- Move `repo` into `d1` to create a chain: repo → d1/repo.
- Try `--rm d1/repo`. This should remove only the moved path, preserving the root.
- No `-f` is needed because removing a chain path is always allowed.

```go
import (
	"fmt"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	repo := filepath.Join(req.WorkRoot, "repo")
	d1 := filepath.Join(req.WorkRoot, "d1")
	mkdirAll(t, repo)
	mkdirAll(t, d1)

	// add repo
	req.Args = []string{"--add", repo}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("add: %s", resp.Output)
	}

	// move repo -> d1/repo (creates chain: repo -> d1/repo)
	req.Args = []string{repo, d1}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move: %s", resp.Output)
	}

	// remove the moved path from the chain
	movedPath := filepath.Join(d1, "repo")
	req.Args = []string{"--rm", movedPath}
	return nil
}
```
