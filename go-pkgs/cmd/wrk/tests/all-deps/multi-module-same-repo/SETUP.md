# Scenario

**Feature**: wrk --all-deps dedups to one worktree per repo when multiple required modules live in the same repo

```
# consumer requires dep1+dep2; myrepo has no root module, two sub-modules svc-a=example.com/dep1, svc-b=example.com/dep2
scan-root (myrepo: svc-a=example.com/dep1, svc-b=example.com/dep2, no root module) + consumer requires dep1+dep2 -> wrk --all-deps -> ONE external worktree of myrepo + two replaces (one per sub-module) + wrk 2 deps
```

## Steps

1. Create a scan-root holding `myrepo`: a git repo with **no root module** and two nested sub-modules `svc-a` (`example.com/dep1`) and `svc-b` (`example.com/dep2`).
2. Create a consumer git repo requiring `example.com/dep1` and `example.com/dep2`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initMultiModuleRepo(t, filepath.Join(scanRoot, "myrepo"))

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1", "example.com/dep2"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}

// initMultiModuleRepo creates a git repo on main at path with NO root go.mod
// and two nested sub-modules: svc-a (example.com/dep1) and svc-b (example.com/dep2).
func initMultiModuleRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	mkdirAll(t, filepath.Join(path, "svc-a"))
	mkdirAll(t, filepath.Join(path, "svc-b"))
	writeFile(t, filepath.Join(path, "svc-a", "go.mod"), "module example.com/dep1\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, "svc-a", "a.go"), "package dep1\n")
	writeFile(t, filepath.Join(path, "svc-b", "go.mod"), "module example.com/dep2\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, "svc-b", "b.go"), "package dep2\n")
	runGit(t, path, "add", ".")
	runGit(t, path, "commit", "-m", "init myrepo with two sub-modules")
}

// nestedExternalAbsSubPath returns the resolved absolute path to a sub-module
// directory inside the external worktree of repoBasename. consumerTop is
// EvalSymlinks-resolved to match wrk's ShowToplevel-resolved output on macOS.
func nestedExternalAbsSubPath(consumerTop, repoBasename, subdir string) string {
	resolved, err := filepath.EvalSymlinks(consumerTop)
	if err != nil {
		resolved = consumerTop
	}
	return filepath.Join(resolved, "external", fmt.Sprintf("%s-main-%s", repoBasename, wrkDate), subdir)
}

// nestedExternalRelSubPath returns the relative ./external/.../<subdir> form
// printed by wrk for a nested sub-module.
func nestedExternalRelSubPath(repoBasename, subdir string) string {
	return fmt.Sprintf("./external/%s-main-%s/%s", repoBasename, wrkDate, subdir)
}
```
