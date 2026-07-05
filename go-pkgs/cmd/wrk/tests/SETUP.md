# Scenario

**Feature**: wrk CLI auto worktree, merge-back, and worktree listing

```
# isolated WRK_HOME + work root per test; build wrk once
wrk (no args) from cwd -> stdout path only + git worktree side effects
wrk --done [--confirm-from-stdin] from linked wt -> merge-back --rm
wrk --list from cwd -> git worktree list stdout unchanged
```

## Preconditions

- The wrk Go module is located one level above the test tree root (at `go-pkgs/cmd/wrk/`)
- Go toolchain is available on PATH
- Git is required for worktree tests

## Context

Each test runs the `wrk` CLI in an isolated environment. The binary is built once and reused across all tests. Each leaf gets its own temp directory and isolated `WRK_HOME` at `{WorkRoot}/.wrk`.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"unicode"

	"github.com/xhd2015/doctest/assert"
)

var buildOnce sync.Once
var builtWrkBin string
var buildWrkErr error

func findModuleRoot(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func getWrkBin(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		modRoot := filepath.Dir(filepath.Dir(DOCTEST_ROOT))
		if modRoot == "" {
			modRoot = findModuleRoot(DOCTEST_ROOT)
		}
		tmpDir, err := os.MkdirTemp("", "wrk-doc-test")
		if err != nil {
			buildWrkErr = fmt.Errorf("create temp dir: %w", err)
			return
		}
		bin := filepath.Join(tmpDir, "wrk")
		cmd := exec.Command("go", "build", "-a", "-o", bin, "./wrk")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildWrkErr = fmt.Errorf("build wrk: %w\n%s", err, out)
			return
		}
		builtWrkBin = bin
	})
	if buildWrkErr != nil {
		t.Fatal(buildWrkErr)
	}
	return builtWrkBin
}

func Setup(t *testing.T, req *Request) error {
	// Resolve symlinks so derived paths match git's resolved output (macOS
	// serves /var from /private/var; t.TempDir returns the symlinked form).
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	ensureHelpersUsed()
	return os.MkdirAll(req.WrkHome, 0755)
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func gitWorktreeList(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "worktree", "list")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("git worktree list in %s: %v\n%s", dir, err, ee.Stderr)
		}
		t.Fatalf("git worktree list in %s: %v", dir, err)
	}
	return string(out)
}

func initGitRepoOnMain(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGit(t, path, "init", "-b", "main")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# test\n")
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", "init")
}

const wrkDate = "2026-06-30"

func sanitizeBranchToken(branch string) string {
	return strings.ReplaceAll(branch, "/", "-")
}

func worktreePath(wrkHome, basename, token, date string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", basename, token, date)
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(wrkHome, "worktrees", name)
}

func branchName(baseBranch, date string, suffix int) string {
	name := baseBranch + "-" + date
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return name
}

func runWrkFrom(t *testing.T, req *Request, dir string) string {
	t.Helper()
	return runWrkWithArgs(t, req, dir)
}

func runWrkWithArgs(t *testing.T, req *Request, dir string, args ...string) string {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
	out, err := cmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("wrk %v exit %d stderr=%q", args, ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("wrk %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s should exist", path)
	}
}

func assertGitFileIsWorktreeLink(t *testing.T, wtDir string) {
	t.Helper()
	gitPath := filepath.Join(wtDir, ".git")
	info, err := os.Stat(gitPath)
	if os.IsNotExist(err) {
		t.Fatalf("%s should exist", gitPath)
	}
	if err != nil {
		t.Fatalf("stat %s: %v", gitPath, err)
	}
	if info.IsDir() {
		t.Fatalf("expected %s to be a regular file (linked worktree), got directory", gitPath)
	}
}

func assertBranchExists(t *testing.T, repoDir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--verify", "refs/heads/"+branch)
	if err := cmd.Run(); err != nil {
		t.Fatalf("branch %q should exist in %s", branch, repoDir)
	}
}

func assertWorktreeListContains(t *testing.T, repoDir, wantPath string) {
	t.Helper()
	list := gitOutput(t, repoDir, "worktree", "list", "--porcelain")
	found := false
	for _, line := range strings.Split(list, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			p := strings.TrimPrefix(line, "worktree ")
			if p == wantPath {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("git worktree list should contain %q, got:\n%s", wantPath, list)
	}
}

func assertBranchCheckedOutInWorktree(t *testing.T, wtDir, wantBranch string) {
	t.Helper()
	got := gitOutput(t, wtDir, "rev-parse", "--abbrev-ref", "HEAD")
	if got != wantBranch {
		t.Fatalf("worktree %s: expected branch %q, got %q", wtDir, wantBranch, got)
	}
}

func v2StdoutTemplate(body string) string {
	if body == "" {
		return "---\nversion: 2\n---\n"
	}
	if !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return "---\nversion: 2\n---\n" + body
}

func joinStdoutBlocks(blocks ...string) string {
	trimmed := make([]string, 0, len(blocks))
	for _, b := range blocks {
		b = strings.TrimSuffix(b, "\n")
		if b != "" {
			trimmed = append(trimmed, b)
		}
	}
	result := strings.Join(trimmed, "\n\n")
	if result != "" && !strings.HasSuffix(result, "\n") {
		result += "\n"
	}
	return result
}

func assertStdoutExactPath(t *testing.T, stdout, wantPath string) {
	t.Helper()
	assert.Output(t, stdout, v2StdoutTemplate(wantPath))
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in %q", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected %q not in %q", substr, s)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	}
}

func assertBranchNotExists(t *testing.T, repoDir, branch string) {
	t.Helper()
	cmd := exec.Command("git", "-C", repoDir, "rev-parse", "--verify", "refs/heads/"+branch)
	if err := cmd.Run(); err == nil {
		t.Fatalf("branch %q should not exist in %s", branch, repoDir)
	}
}

func assertWorktreeListNotContains(t *testing.T, repoDir, wantPath string) {
	t.Helper()
	list := gitOutput(t, repoDir, "worktree", "list", "--porcelain")
	for _, line := range strings.Split(list, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			p := strings.TrimPrefix(line, "worktree ")
			if p == wantPath {
				t.Fatalf("git worktree list should not contain %q", wantPath)
			}
		}
	}
}

// setupWrkWorktreeFromMain creates myrepo on main, runs wrk once, and records paths on req.
func setupWrkWorktreeFromMain(t *testing.T, req *Request) (mainRepo, wtDir, branch string) {
	t.Helper()
	mainRepo = filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	// wrk --done requires a consumer go.mod (it tidies after merge-back).
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGit(t, mainRepo, "add", "go.mod")
	runGit(t, mainRepo, "commit", "-m", "add go.mod")
	wtDir = runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	branch = branchName("main", wrkDate, 0)
	req.WtBranch = branch
	return mainRepo, wtDir, branch
}

func commitAheadOnWorktree(t *testing.T, wtDir, filename, content string) {
	t.Helper()
	writeFile(t, filepath.Join(wtDir, filename), content)
	runGit(t, wtDir, "add", filename)
	runGit(t, wtDir, "commit", "-m", "worktree commit")
}

// slugify converts a task description into a path-safe slug.
// Rules: lowercase, non-letter-non-digit → "-", collapse runs of "-",
// trim leading/trailing "-", truncate to 64 runes.
func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	s = b.String()
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	runes := []rune(s)
	if len(runes) > 64 {
		s = string(runes[:64])
	}
	s = strings.Trim(s, "-")
	return s
}

func worktreePathWithTask(wrkHome, basename, token, date, slug string, suffix int) string {
	name := fmt.Sprintf("%s-%s-%s", basename, token, date)
	if slug != "" {
		name = fmt.Sprintf("%s-%s", name, slug)
	}
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return filepath.Join(wrkHome, "worktrees", name)
}

func branchNameWithTask(baseBranch, date, slug string, suffix int) string {
	name := baseBranch + "-" + date
	if slug != "" {
		name = name + "-" + slug
	}
	if suffix > 0 {
		name = fmt.Sprintf("%s-%d", name, suffix)
	}
	return name
}

// createWorktreeWithTask is like setupWrkWorktreeFromMain but with a task description.
func wrkEnv(req *Request) []string {
	if req.UseMinimalPath {
		home := req.FakeHome
		if home == "" {
			home = req.WorkRoot
		}
		env := []string{
			"HOME=" + home,
			"PATH=/usr/bin:/bin",
			"WRK_HOME=" + req.WrkHome,
			"WRK_DATE=" + wrkDate,
		}
		if req.SetTaskEnv != "" {
			env = append(env, req.SetTaskEnv)
		}
		if req.BasenameEnv != "" {
			env = append(env, req.BasenameEnv)
		}
		if req.ProjectsPerfLog != "" {
			env = append(env, "WRK_PROJECTS_PERF_LOG="+req.ProjectsPerfLog)
		}
		return env
	}
	env := append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
	if req.FakeHome != "" {
		env = append(env, "HOME="+req.FakeHome)
	}
	if req.SetTaskEnv != "" {
		env = append(env, req.SetTaskEnv)
	}
	if req.BasenameEnv != "" {
		env = append(env, req.BasenameEnv)
	}
	if req.ProjectsPerfLog != "" {
		env = append(env, "WRK_PROJECTS_PERF_LOG="+req.ProjectsPerfLog)
	}
	return env
}

func createWorktreeWithTask(t *testing.T, req *Request, taskDesc string) (mainRepo, wtDir, branch string) {
	t.Helper()
	mainRepo = filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\ngo 1.21\n")
	runGit(t, mainRepo, "add", "go.mod")
	runGit(t, mainRepo, "commit", "-m", "add go.mod")
	slug := slugify(taskDesc)
	wtDir = runWrkWithArgs(t, req, mainRepo, "--task", taskDesc)
	req.WtDir = wtDir
	branch = branchNameWithTask("main", wrkDate, slug, 0)
	req.WtBranch = branch
	req.TaskDesc = taskDesc
	return mainRepo, wtDir, branch
}

func ensureHelpersUsed() {
	_ = mkdirAll
	_ = writeFile
	_ = skipIfNoGit
	_ = runGit
	_ = gitOutput
	_ = gitWorktreeList
	_ = initGitRepoOnMain
	_ = sanitizeBranchToken
	_ = worktreePath
	_ = branchName
	_ = runWrkFrom
	_ = runWrkWithArgs
	_ = assertErrIsNil
	_ = assertFileExists
	_ = assertGitFileIsWorktreeLink
	_ = assertBranchExists
	_ = assertWorktreeListContains
	_ = assertBranchCheckedOutInWorktree
	_ = v2StdoutTemplate
	_ = joinStdoutBlocks
	_ = assertStdoutExactPath
	_ = assertContains
	_ = assertNotContains
	_ = assertFileNotExists
	_ = assertBranchNotExists
	_ = assertWorktreeListNotContains
	_ = setupWrkWorktreeFromMain
	_ = commitAheadOnWorktree
	_ = slugify
	_ = worktreePathWithTask
	_ = branchNameWithTask
	_ = createWorktreeWithTask
	_ = wrkEnv
}
```