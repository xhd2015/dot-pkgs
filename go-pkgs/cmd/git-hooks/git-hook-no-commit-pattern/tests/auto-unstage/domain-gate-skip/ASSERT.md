## Expected

- The command exits successfully (exit 0) because the domain gate skipped.
- No output (the hook didn't scan staged files).
- `main.go` is still in the staging area (nothing was checked or unstaged).

## Exit Code

- Exit code is `0`.

```go
import "github.com/xhd2015/doctest/session"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for domain gate skip, got %d:\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "" {
		t.Fatalf("expected no output for domain gate skip, got:\n%s", resp.Output)
	}
	staged, err := getStagedFileNames(req.RepoDir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(staged, "main.go") {
		t.Fatal("expected main.go to still be staged (hook was skipped), but it was unstaged")
	}
}
```
