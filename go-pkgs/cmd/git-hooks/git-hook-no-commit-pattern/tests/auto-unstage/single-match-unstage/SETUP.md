# Scenario

**Feature**: `--auto-unstage` with a single match unstages the file and exits 0

```
# stage main.go, run with --auto-unstage "*.go" -> print main.go, unstage it, exit 0
git add main.go -> hook --auto-unstage "*.go" -> main.go matches -> print "main.go" -> restore --staged main.go -> exit 0
```

## Preconditions

- Repository has an initial commit.
- `main.go` is staged.

## Steps

1. Create an initial commit.
2. Create and stage `main.go`.
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
	req.Args = []string{"--auto-unstage", "*.go"}
	return nil
}
```
