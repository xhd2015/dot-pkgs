## Expected

- The command exits successfully (exit 0).
- The output contains `main.go`.
- `main.go` is no longer in the staging area.

## Side Effects

- `main.go` is removed from the git index.
- `main.go` still exists in the working tree.

## Exit Code

- Exit code is `0`.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 with --auto-unstage, got %d:\n%s", resp.ExitCode, resp.Output)
	}
	if !strings.Contains(resp.Output, "main.go") {
		t.Fatalf("expected output to contain main.go, got:\n%s", resp.Output)
	}
	staged, err := getStagedFileNames(req.RepoDir)
	if err != nil {
		t.Fatal(err)
	}
	if containsString(staged, "main.go") {
		t.Fatal("expected main.go to be unstaged, but it is still in the index")
	}
}
```
