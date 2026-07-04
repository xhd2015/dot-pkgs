## Expected

- The command exits successfully (exit 0).
- The output contains `go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md`.
- The requirement doc is no longer in the staging area.

## Side Effects

- `go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md` is removed from the git index.
- The file still exists in the working tree.

## Exit Code

- Exit code is `0`.

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	want := "go-pkgs/REQUIREMENT-DESIGN-wrk-status-compare.md"
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 with --auto-unstage for subdirectory REQUIREMENT-*.md match, got %d:\n%s", resp.ExitCode, resp.Output)
	}
	if !strings.Contains(resp.Output, want) {
		t.Fatalf("expected output to contain %q, got:\n%s", want, resp.Output)
	}
	staged, err := getStagedFileNames(req.RepoDir)
	if err != nil {
		t.Fatal(err)
	}
	if containsString(staged, want) {
		t.Fatalf("expected %q to be unstaged, but it is still in the index: %v", want, staged)
	}
}
```