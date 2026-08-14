# Scenario

**Feature**: valid repo + SHA + at least one change flag

```
# two-commit master repo; target is HEAD
initRepo -> commit init -> commit "fix typo" -> req.Args = -C dir <sha>
```

## Preconditions

- Isolated `t.TempDir()` repo on branch `master`.
- Parent commit `init` plus target commit `fix typo` so `HEAD^` exists.
- Author/committer identities come from helper `cmd.Env` (see root SETUP).

## Steps

1. Create the repo and two commits.
2. Fill `req` commit metadata and set `req.Args` to `-C <dir> <oldsha>`.
3. Descendants append change flags, topology, remotes, or `--dry-run`.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.Dir = initRepo(t)
	commitFile(t, req.Dir, "README", "init\n", "init")
	commitFile(t, req.Dir, "note.txt", "v1\n", "fix typo")
	fillCommitMeta(t, req)
	return nil
}
```
