# Scenario

**Feature**: `--auto-unstage` with no matching files exits 0 and prints nothing

```
# stage main.go, run with --auto-unstage "*.md" -> no match -> exit 0, no output
git add main.go -> hook --auto-unstage "*.md" -> main.go does not match -> exit 0, nothing printed, nothing unstaged
```

## Preconditions

- Repository has an initial commit.
- `main.go` is staged but does not match the pattern.

## Steps

1. Create an initial commit.
2. Create and stage `main.go`.
3. Run the hook with `--auto-unstage *.md` (no match).

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	if err := initGitRepoWithCommit(req.RepoDir); err != nil {
		return err
	}
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	req.Args = []string{"--auto-unstage", "*.md"}
	return nil
}
```
