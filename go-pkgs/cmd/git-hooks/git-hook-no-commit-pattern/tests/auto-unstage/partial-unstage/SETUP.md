# Scenario

**Feature**: `--auto-unstage` only unstages matched files, leaving non-matched files staged

```
# stage main.go README.md, run with --auto-unstage "*.go" -> main.go matches, README.md does not
git add main.go README.md -> hook --auto-unstage "*.go" -> main.go matches -> unstage main.go -> README.md stays staged -> exit 0
```

## Preconditions

- Repository has an initial commit.
- `main.go` (matches `*.go`) and `README.md` (does not match) are staged.

## Steps

1. Create an initial commit.
2. Create and stage `main.go` and `README.md`.
3. Run the hook with `--auto-unstage *.go`.

```go
func Setup(t *testing.T, req *Request) error {
	if err := initGitRepoWithCommit(req.RepoDir); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "README.md", "# Readme\n"); err != nil {
		return err
	}
	req.Args = []string{"--auto-unstage", "*.go"}
	return nil
}
```
