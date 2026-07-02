## Expected

- The command exits successfully (exit 0).
- No output (nothing matched).
- `main.go` is still in the staging area (nothing was unstaged).

## Exit Code

- Exit code is `0`.

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for no match, got %d:\n%s", resp.ExitCode, resp.Output)
	}
	if resp.Output != "" {
		t.Fatalf("expected no output for no match, got:\n%s", resp.Output)
	}
	staged, err := getStagedFileNames(req.RepoDir)
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(staged, "main.go") {
		t.Fatal("expected main.go to still be staged, but it was unstaged")
	}
}
```
