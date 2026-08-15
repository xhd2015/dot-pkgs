## Expected

- The command exits successfully (exit 0).
- The output contains both `main.go` and `test.go`.
- Neither file is in the staging area.

## Side Effects

- Both matched files are removed from the git index.
- Both files still exist in the working tree.

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
	for _, want := range []string{"main.go", "test.go"} {
		if !strings.Contains(resp.Output, want) {
			t.Fatalf("expected output to contain %q, got:\n%s", want, resp.Output)
		}
	}
	staged, err := getStagedFileNames(req.RepoDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"main.go", "test.go"} {
		if containsString(staged, name) {
			t.Fatalf("expected %s to be unstaged, but it is still in the index", name)
		}
	}
}
```
