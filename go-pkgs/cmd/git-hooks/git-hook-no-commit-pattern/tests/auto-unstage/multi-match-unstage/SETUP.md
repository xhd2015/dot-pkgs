# Scenario

**Feature**: `--auto-unstage` with multiple matches unstages all matched files

```
# stage main.go test.go, run with --auto-unstage "*.go" -> print both, unstage both, exit 0
git add main.go test.go -> hook --auto-unstage "*.go" -> both match -> print both -> restore both -> exit 0
```

## Preconditions

- Repository has an initial commit.
- `main.go` and `test.go` are staged.

## Steps

1. Create an initial commit.
2. Create and stage `main.go` and `test.go`.
3. Run the hook with `--auto-unstage *.go`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := initGitRepoWithCommit(req.RepoDir); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "test.go", "package test\n"); err != nil {
		return err
	}
	req.Args = []string{"--auto-unstage", "*.go"}
	return nil
}
```
