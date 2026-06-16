# Scenario

--back skips worktree entries to find the correct previous location. Was a bug: used to try moving onto the worktree dir.

mvd -w base target → [(base), (target w:target)]
mvd base another → [(base), (target w:target), (another)]
mvd --back another → [(base), (target w:target)]

## Steps
- Create a git repo at work/base with one commit.
- Use `mvd -w base target` to create a worktree at work/target.
- Use `mvd base another` to move the main repo to work/another.
- Use `mvd --back another` to move the repo back to work/base.

The --back should skip the worktree entry in the chain and move another back to base.
Before the fix, this fails because --back tries to move another onto target (which still exists).

```go
import (
	"fmt"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	repo := filepath.Join(req.WorkRoot, "base")
	mkdirAll(t, repo)
	initGitRepo(t, repo)

	// Step 1: create worktree from base
	wt := filepath.Join(req.WorkRoot, "target")
	req.Args = []string{"-w", repo, wt}
	resp, err := runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("worktree create failed: %s", resp.Output)
	}

	// Step 2: plain move base → another
	another := filepath.Join(req.WorkRoot, "another")
	req.Args = []string{repo, another}
	resp, err = runMvd(t, req)
	if err != nil {
		return err
	}
	if resp.ExitCode != 0 {
		return fmt.Errorf("move base→another failed: %s", resp.Output)
	}

	// Step 3: --back another → should move back to base
	req.Args = []string{"--back", another}
	return nil
}
```
