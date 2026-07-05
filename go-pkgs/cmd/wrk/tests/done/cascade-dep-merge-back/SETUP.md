# Scenario

**Feature**: wrk --done cascade merges an ahead external dep worktree back into the dep repo

```
# external dep wt is a worktree of the DEP repo; a dep fix committed on it is
# ahead of dep main. --done cascades MergeBack over external/* first, ff-merging
# the dep branch into the dep repo, then removes the worktree.
consumer wt + external/dep wt (ahead) -> wrk --done --confirm-from-stdin -> dep branch merged into dep main, ext wt removed
```

## Steps

1. Create consumer main repo with `go.mod` requiring `example.com/dep`.
2. `wrk` creates the consumer linked worktree.
3. `wrk --dep <depRepo>` spawns `external/mydep-main-{date}` (a worktree of the dep repo).
4. Commit a dep fix on the external worktree (ahead of dep main).
5. Run `wrk --done --confirm-from-stdin` from the consumer worktree, piping `Y`.

## Expected (correct) behavior

The cascade runs before the consumer's local-replace guard. For the ahead dep
worktree, `MergeBack` resolves the owning main repo from the worktree's `.git`
gitdir (the dep main) and ff-merges the dep branch into the dep repo — so the
dep fix lands in dep main — then removes the external worktree + branch. The
consumer still carries `replace => ./external/...`, so the guard then blocks the
consumer's own merge-back (non-zero exit, consumer worktree remains). This test
pins the **merge-back** semantics: ahead dep work is not discarded.

```go
import (
	"os/exec"
	"path/filepath"
)

const cascadeMergeBackDepModule = "example.com/dep"

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoMod(t, mainRepo, "edit", "-require="+cascadeMergeBackDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+cascadeMergeBackDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	req.ExternalWtDir = externalPath

	// Commit a dep fix on the external worktree → ahead of dep main.
	writeFile(t, filepath.Join(externalPath, "dep.go"), "package dep // fix\n")
	runGitIsolated(t, externalPath, "add", "dep.go")
	runGitIsolated(t, externalPath, "commit", "-m", "dep fix on worktree")

	req.RepoDir = wtDir
	req.Args = []string{"--done", "--confirm-from-stdin"}
	req.StdinInput = "\n"
	return nil
}

func runGoMod(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```
