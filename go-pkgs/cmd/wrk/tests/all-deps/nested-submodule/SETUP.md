# Scenario

**Feature**: wrk --all-deps links a required module nested inside a larger repo via mod/scan sub-module discovery

```
# consumer requires example.com/dep; myrepo root module is example.com/myrepo (not required), nested services/dep is example.com/dep (required)
scan-root (myrepo: root=example.com/myrepo, services/dep=example.com/dep) + consumer requires example.com/dep -> wrk --all-deps -> external worktree of myrepo + replace example.com/dep => external/myrepo-main-{date}/services/dep + wrk 1 deps
```

## Steps

1. Create a scan-root holding `myrepo`: a git repo whose **root** module is `example.com/myrepo` (not required by the consumer) and whose **nested** sub-module at `services/dep` is `example.com/dep` (required).
2. Create a consumer git repo requiring `example.com/dep`.
3. Run `wrk --all-deps --scan-root <scanRoot>` from the consumer.

```go
func Setup(t *testing.T, req *Request) error {
	allDepsEnsureHelpersUsed()

	scanRoot := filepath.Join(req.WorkRoot, "scan-root")
	mkdirAll(t, scanRoot)
	initNestedSubmoduleRepo(t, filepath.Join(scanRoot, "myrepo"))

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep"}, "")

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.Args = []string{"--all-deps", "--scan-root", scanRoot}
	return nil
}

// initNestedSubmoduleRepo creates a git repo on main at path whose root module
// is example.com/myrepo (with a root .go file) and whose nested sub-module at
// services/dep is example.com/dep. The root module is intentionally NOT the one
// the consumer requires, so the nested sub-module is the only match.
func initNestedSubmoduleRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "go.mod"), "module example.com/myrepo\n\ngo 1.22\n")
	writeFile(t, filepath.Join(path, "root.go"), "package myrepo\n")
	depDir := filepath.Join(path, "services", "dep")
	mkdirAll(t, depDir)
	writeFile(t, filepath.Join(depDir, "go.mod"), "module example.com/dep\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depDir, "dep.go"), "package dep\n")
	runGit(t, path, "add", ".")
	runGit(t, path, "commit", "-m", "init myrepo with nested dep")
}

// nestedExternalAbsSubPath returns the resolved absolute path to a sub-module
// directory inside the external worktree of repoBasename, e.g.
// <consumerTop>/external/myrepo-main-{date}/services/dep. consumerTop is
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

func nestedEnsureHelpersUsed() {
	_ = initNestedSubmoduleRepo
	_ = nestedExternalAbsSubPath
	_ = nestedExternalRelSubPath
}
```
