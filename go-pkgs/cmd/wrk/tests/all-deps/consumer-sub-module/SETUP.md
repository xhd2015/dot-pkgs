# Scenario

**Feature**: wrk --all-deps works when consumer module lives in a subdirectory (no root go.mod)

```
# consumer go.mod in go-pkgs/ requires dep1+dep2; scan-root has both repos -> wrk --all-deps links both
consumer (go-pkgs/ requires dep1, dep2) + scan-root (dep1, dep2) -> wrk --all-deps -> 2 deps
```

## Preconditions

- Git and Go must be available.
- Consumer repo root has NO go.mod; `go-pkgs/go.mod` requires `example.com/dep1` and `example.com/dep2`.
- Consumer cwd is the repo root.

## Steps

1. Create scan-root with dep repos `mydep1` (`example.com/dep1`) and `mydep2` (`example.com/dep2`).
2. Create consumer git repo with `go-pkgs/go.mod` requiring both deps; no `go.mod` at root.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer repo root.

```go
import "path/filepath"

func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep1"), "example.com/dep1", "dep1")
	initAllDepsRepo(t, filepath.Join(scanRoot, "mydep2"), "example.com/dep2", "dep2")

	consumer := filepath.Join(req.WorkRoot, "consumer")
	mkdirAll(t, consumer)
	runGit(t, consumer, "init", "-b", "main")
	runGit(t, consumer, "config", "user.email", "test@test.com")
	runGit(t, consumer, "config", "user.name", "Test")
	// NO go.mod at repo root. Module lives in go-pkgs/.
	modDir := filepath.Join(consumer, "go-pkgs")
	mkdirAll(t, modDir)
	writeFile(t, filepath.Join(modDir, "go.mod"), "module example.com/consumer\n\ngo 1.22\n\nrequire example.com/dep1 v0.0.0\nrequire example.com/dep2 v0.0.0\n")
	writeFile(t, filepath.Join(modDir, "main.go"), "package main\n")
	runGit(t, consumer, "add", ".")
	runGit(t, consumer, "commit", "-m", "init consumer")

	consumer, err := filepath.EvalSymlinks(consumer)
	if err != nil {
		t.Fatalf("eval symlinks %s: %v", consumer, err)
	}
	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```