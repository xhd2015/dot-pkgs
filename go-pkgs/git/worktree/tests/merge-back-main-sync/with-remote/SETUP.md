# Scenario

**Feature**: origin remote present → main-sync applies

```
bare origin + tracking master -> sync before land
```

## Steps

1. Init main on `master`, add bare `origin`, push `-u`.
2. Add feature worktree with a unique commit.
3. Advance origin ahead of main via a second clone (leaves main behind).

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func setupWithRemoteBehind(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "main")
	if err := os.MkdirAll(mainRepo, 0755); err != nil {
		t.Fatal(err)
	}
	initRepo(t, mainRepo, "master")

	bare := filepath.Join(req.WorkRoot, "bare.git")
	runGit(t, req.WorkRoot, "clone", "--bare", mainRepo, bare)
	runGit(t, mainRepo, "remote", "add", "origin", bare)
	runGit(t, mainRepo, "push", "-u", "origin", "master")

	featureWT := filepath.Join(req.WorkRoot, "feature")
	addWorktree(t, mainRepo, featureWT, "feature")
	if err := os.WriteFile(filepath.Join(featureWT, "feature.txt"), []byte("f\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, featureWT, "add", "feature.txt")
	runGit(t, featureWT, "commit", "-m", "feature")

	// Advance origin without updating main.
	other := filepath.Join(req.WorkRoot, "other")
	runGit(t, req.WorkRoot, "clone", bare, other)
	runGit(t, other, "config", "user.email", "test@test.com")
	runGit(t, other, "config", "user.name", "Test")
	if err := os.WriteFile(filepath.Join(other, "remote-only.txt"), []byte("r\n"), 0644); err != nil {
		t.Fatal(err)
	}
	runGit(t, other, "add", "remote-only.txt")
	runGit(t, other, "commit", "-m", "remote only")
	runGit(t, other, "push", "origin", "master")

	req.MainRepo = mainRepo
	req.SourcePath = featureWT
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	setupWithRemoteBehind(t, req)
	return nil
}
```
