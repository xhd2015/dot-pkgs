## Expected

- `with-git` entry: `Aggregates.GitRepos == 1`.
- `plain-dir` entry: `Aggregates.GitRepos == 0`.
- Summary `GitRepos == 1`.

## Errors

- `err` is nil.
- Git entry missing repo count.

```go
import (
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if req.Home == "" {
		t.Skip("git not available; Setup skipped seeding")
	}
	if err != nil {
		t.Fatal(err)
	}

	withGit := findEntry(t, resp.Entries, "with-git")
	if withGit.Aggregates.GitRepos != 1 {
		t.Fatalf("with-git GitRepos = %d, want 1", withGit.Aggregates.GitRepos)
	}

	plain := findEntry(t, resp.Entries, "plain-dir")
	if plain.Aggregates.GitRepos != 0 {
		t.Fatalf("plain-dir GitRepos = %d, want 0", plain.Aggregates.GitRepos)
	}

	if resp.Summary.GitRepos != 1 {
		t.Fatalf("summary.GitRepos = %d, want 1", resp.Summary.GitRepos)
	}
}
```