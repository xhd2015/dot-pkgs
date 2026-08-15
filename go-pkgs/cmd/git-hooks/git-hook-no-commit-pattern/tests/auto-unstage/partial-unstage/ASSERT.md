## Expected

- The command exits successfully (exit 0).
- The output contains `main.go` (matched).
- The output does NOT contain `README.md` (not matched).
- `main.go` is no longer in the staging area.
- `README.md` is still in the staging area.

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
	if strings.Contains(resp.Output, "README.md") {
		t.Fatalf("expected output NOT to contain README.md, got:\n%s", resp.Output)
	}
	staged, err := getStagedFileNames(req.RepoDir)
	if err != nil {
		t.Fatal(err)
	}
	if containsString(staged, "main.go") {
		t.Fatal("expected main.go to be unstaged, but it is still in the index")
	}
	if !containsString(staged, "README.md") {
		t.Fatal("expected README.md to still be staged, but it was also unstaged")
	}
}
```
