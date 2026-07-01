# Scenario

**Feature**: wrk --all-deps errors when cwd is not a git repository

```
# cwd is a plain directory (no .git) -> wrk --all-deps --scan-root <tmp> -> non-zero, is not a git repository
plain dir (no .git) -> wrk --all-deps --scan-root <tmp> -> error (is not a git repository)
```

## Steps

1. Create a non-git temp dir as the cwd.
2. Create an empty scan-root temp dir.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the non-git cwd.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	// cwd is a plain directory, intentionally NOT a git repo.
	req.RepoDir = t.TempDir()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)

	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}
```
