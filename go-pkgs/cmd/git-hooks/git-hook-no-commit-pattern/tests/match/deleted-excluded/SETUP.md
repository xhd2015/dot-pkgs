# Scenario

**Feature**: deleted staged files are excluded from pattern matching

```
# stage main.go, delete main.go, pattern "*.go" -> deleted excluded -> exit 0
stage main.go -> git rm main.go -> pattern "*.go" -> deleted excluded -> exit 0
```

## Preconditions

- A file must first be committed to the repository so it can be deleted and staged.

## Steps

1. Create, commit, then delete `main.go` and stage the deletion.
2. Run the hook with pattern `*.go`.

```go
import "github.com/xhd2015/doctest/session"

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"*.go"}
	// Create and commit main.go so we can delete it
	if err := writeAndStage(req.RepoDir, "main.go", "package main\n"); err != nil {
		return err
	}
	if err := runGit(req.RepoDir, "commit", "--no-verify", "-m", "initial"); err != nil {
		return err
	}
	// Delete and stage the deletion
	if err := deleteAndStage(req.RepoDir, "main.go"); err != nil {
		return err
	}
	return nil
}
```
