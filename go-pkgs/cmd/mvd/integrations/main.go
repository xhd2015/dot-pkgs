package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// --- test framework ---

type T struct {
	name   string
	failed bool
	logs   []string
}

func (t *T) Logf(format string, args ...interface{}) {
	t.logs = append(t.logs, fmt.Sprintf(format, args...))
}

func (t *T) Errorf(format string, args ...interface{}) {
	t.failed = true
	msg := fmt.Sprintf(format, args...)
	t.logs = append(t.logs, "ERROR: "+msg)
}

func (t *T) Fatalf(format string, args ...interface{}) {
	t.Errorf(format, args...)
	panic(testFail{t})
}

func (t *T) Skipf(format string, args ...interface{}) {
	panic(testSkip{fmt.Sprintf(format, args...)})
}

func (t *T) Helper() {}

type testFail struct{ t *T }
type testSkip struct{ reason string }

type testCase struct {
	name string
	fn   func(t *T)
}

var (
	mvdBin      string
	tmpBase     string
	passed      int
	failed      int
	skipped     int
	failDetails []string
)

func main() {
	var err error
	mvdBin, err = buildMvd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "build mvd: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(mvdBin)

	tmpBase, err = os.MkdirTemp("", "mvd-integ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "create tmp base: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpBase)

	tests := []testCase{
		{"TestBasicMove", testBasicMove},
		{"TestMoveToExistingDir", testMoveToExistingDir},
		{"TestMultiStepMove", testMultiStepMove},
		{"TestMoveAsRootPath", testMoveAsRootPath},
		{"TestMoveByBasename", testMoveByBasename},
		{"TestMoveByAlias", testMoveByAlias},
		{"TestAmbiguousBasename", testAmbiguousBasename},
		{"TestAdd", testAdd},
		{"TestAddDuplicate", testAddDuplicate},
		{"TestRemove", testRemove},
		{"TestRemoveForce", testRemoveForce},
		{"TestRemoveNoForceWithHistory", testRemoveNoForceWithHistory},
		{"TestRebase", testRebase},
		{"TestBack", testBack},
		{"TestBackAtOrigin", testBackAtOrigin},
		{"TestBackByBasename", testBackByBasename},
		{"TestListAll", testListAll},
		{"TestListSingle", testListSingle},
		{"TestClear", testClear},
		{"TestMultiStepBack", testMultiStepBack},
		{"TestNonExistentSrc", testNonExistentSrc},
		{"TestWorktreeMove", testWorktreeMove},
		{"TestWorktreeNonGitSrc", testWorktreeNonGitSrc},
		{"TestWorktreeBackDirty", testWorktreeBackDirty},
		{"TestWorktreeBackUnmerged", testWorktreeBackUnmerged},
		{"TestWorktreeBackSuccess", testWorktreeBackSuccess},
		{"TestWorktreeBranchCollision", testWorktreeBranchCollision},
	}

	for _, tc := range tests {
		runTest(tc)
	}

	fmt.Println()
	if failed > 0 {
		fmt.Printf("%d passed, %d failed", passed, failed)
		if skipped > 0 {
			fmt.Printf(", %d skipped", skipped)
		}
		fmt.Println()
		for _, d := range failDetails {
			fmt.Println(d)
		}
		fmt.Println("FAIL")
		os.Exit(1)
	}
	fmt.Printf("%d passed", passed)
	if skipped > 0 {
		fmt.Printf(", %d skipped", skipped)
	}
	fmt.Println()
	fmt.Println("PASS")
}

func runTest(tc testCase) {
	t := &T{name: tc.name}
	fmt.Printf("=== RUN   %s\n", tc.name)

	defer func() {
		r := recover()
		if r != nil {
			if sf, ok := r.(testFail); ok {
				for _, l := range sf.t.logs {
					fmt.Printf("    %s\n", l)
				}
				fmt.Printf("--- FAIL: %s\n", tc.name)
				failed++
				failDetails = append(failDetails, fmt.Sprintf("  FAIL: %s", tc.name))
			} else if ts, ok := r.(testSkip); ok {
				fmt.Printf("    SKIP: %s\n", ts.reason)
				fmt.Printf("--- SKIP: %s\n", tc.name)
				skipped++
			} else {
				panic(r)
			}
			return
		}
		if t.failed {
			for _, l := range t.logs {
				fmt.Printf("    %s\n", l)
			}
			fmt.Printf("--- FAIL: %s\n", tc.name)
			failed++
			failDetails = append(failDetails, fmt.Sprintf("  FAIL: %s", tc.name))
		} else {
			fmt.Printf("--- PASS: %s\n", tc.name)
			passed++
		}
	}()

	_ = os.MkdirAll(tmpBase, 0755)
	tc.fn(t)
}

// --- helpers ---

func buildMvd() (string, error) {
	modRoot := findModuleRoot()
	if modRoot == "" {
		return "", fmt.Errorf("cannot find module root (no go.mod found)")
	}
	bin := filepath.Join(os.TempDir(), "mvd-integ-test")
	// Build the mvd package relative to the module root
	// The mvd package is in cmd/mvd under go-pkgs, but cmd/ has its own go.mod
	mvdPkg := "."
	// If module root is cmd/, use ./mvd; otherwise use ./cmd/mvd from go-pkgs
	if filepath.Base(modRoot) == "cmd" {
		mvdPkg = "./mvd"
	} else {
		mvdPkg = "./cmd/mvd"
	}
	cmd := exec.Command("go", "build", "-o", bin, mvdPkg)
	cmd.Dir = modRoot
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w\n%s", err, out)
	}
	return bin, nil
}

func findModuleRoot() string {
	// Get the directory of this source file at runtime
	// Start from current working directory and walk up
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if _, err := os.Stat(filepath.Join(d, "go.mod")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return ""
		}
		d = parent
	}
}

func testDir(t *T, name string) string {
	d := filepath.Join(tmpBase, t.name, name)
	t.Helper()
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	return d
}

func configHome(t *T) string {
	d := filepath.Join(tmpBase, t.name, ".mvd-config")
	t.Helper()
	if err := os.MkdirAll(d, 0755); err != nil {
		t.Fatalf("mkdir config home: %v", err)
	}
	return d
}

func runMvd(t *T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command(mvdBin, args...)
	cmd.Env = append(os.Environ(), "MVD_DEBUG_CONFIG_HOME="+configHome(t))
	out, err := cmd.CombinedOutput()
	t.Logf(string(out))
	if err != nil {
		return string(out), fmt.Errorf("%s", strings.TrimSpace(string(out)))
	}
	return string(out), nil
}

func runMvdOk(t *T, args ...string) string {
	out, err := runMvd(t, args...)
	if err != nil {
		t.Fatalf("mvd %v: %v", args, err)
	}
	return out
}

func runMvdErr(t *T, args ...string) string {
	out, err := runMvd(t, args...)
	if err == nil {
		t.Fatalf("mvd %v: expected error, got nil", args)
	}
	return out + err.Error()
}

type historyFile struct {
	Version  string                `json:"version"`
	Projects map[string]projEntry  `json:"projects"`
}

type projEntry struct {
	Locations []locEntry `json:"locations"`
}

type locEntry struct {
	Path string   `json:"path"`
	Git  *gitMeta `json:"git,omitempty"`
}

type gitMeta struct {
	Type     string `json:"type"`
	MainRepo string `json:"main_repo,omitempty"`
	Branch   string `json:"branch,omitempty"`
}

func readHistory(t *T) *historyFile {
	t.Helper()
	p := filepath.Join(configHome(t), "history.json")
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	var h historyFile
	if err := json.Unmarshal(data, &h); err != nil {
		t.Fatalf("parse history: %v", err)
	}
	return &h
}

func readAliases(t *T) map[string]string {
	t.Helper()
	p := filepath.Join(configHome(t), "aliases.json")
	data, err := os.ReadFile(p)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read aliases: %v", err)
	}
	var a map[string]string
	if err := json.Unmarshal(data, &a); err != nil {
		t.Fatalf("parse aliases: %v", err)
	}
	return a
}

func assertHistoryNil(t *T) {
	h := readHistory(t)
	if h != nil && len(h.Projects) > 0 {
		t.Fatalf("expected no history, got projects=%d", len(h.Projects))
	}
}

func assertHistoryLen(t *T, n int) *historyFile {
	h := readHistory(t)
	if h == nil {
		t.Fatalf("expected history, got nil")
	}
	if len(h.Projects) != n {
		t.Fatalf("expected %d projects, got %d", n, len(h.Projects))
	}
	return h
}

func assertHistoryChain(t *T, key string, wantPaths ...string) {
	h := assertHistoryLen(t, 1)
	proj, ok := h.Projects[key]
	if !ok {
		for _, p := range h.Projects {
			if len(p.Locations) > 0 && p.Locations[0].Path == key {
				proj = p
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("project %s not found in history", key)
		}
	}
	if len(proj.Locations) != len(wantPaths) {
		t.Fatalf("expected %d locations, got %d: %+v", len(wantPaths), len(proj.Locations), locPaths(proj.Locations))
	}
	for i, want := range wantPaths {
		if proj.Locations[i].Path != want {
			t.Fatalf("location[%d]: expected %s, got %s", i, want, proj.Locations[i].Path)
		}
	}
}

func assertHistoryContainsKey(t *T, key string) *historyFile {
	h := readHistory(t)
	if h == nil {
		t.Fatalf("expected history, got nil")
	}
	if _, ok := h.Projects[key]; !ok {
		t.Fatalf("project %s not found in history", key)
	}
	return h
}

func assertHistoryNotContainsKey(t *T, key string) {
	h := readHistory(t)
	if h == nil {
		return
	}
	if _, ok := h.Projects[key]; ok {
		t.Fatalf("project %s should not be in history", key)
	}
}

func locPaths(locs []locEntry) []string {
	paths := make([]string, len(locs))
	for i, l := range locs {
		paths[i] = l.Path
	}
	return paths
}

func assertFileExists(t *T, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s should exist", path)
	}
}

func assertFileNotExists(t *T, path string) {
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	}
}

func assertNoErr(t *T, err error) {
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func writeFile(t *T, path, content string) {
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}

func assertContains(t *T, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Fatalf("expected output to contain %q, got:\n%s", substr, s)
	}
}

func assertNotContains(t *T, s, substr string) {
	if strings.Contains(s, substr) {
		t.Fatalf("output should not contain %q, got:\n%s", substr, s)
	}
}

func assertHistoryWorktreeEntry(t *T, key string, idx int, mainRepo, branch string) {
	h := readHistory(t)
	if h == nil {
		t.Fatalf("expected history, got nil")
	}
	proj, ok := h.Projects[key]
	if !ok {
		for _, p := range h.Projects {
			if len(p.Locations) > 0 && p.Locations[0].Path == key {
				proj = p
				ok = true
				break
			}
		}
		if !ok {
			t.Fatalf("project %s not found in history", key)
		}
	}
	if idx >= len(proj.Locations) {
		t.Fatalf("index %d out of range (len=%d)", idx, len(proj.Locations))
	}
	loc := proj.Locations[idx]
	if loc.Git == nil {
		t.Fatalf("location[%d] has no git metadata", idx)
	}
	if loc.Git.Type != "worktree" {
		t.Fatalf("location[%d] git.type: expected 'worktree', got %q", idx, loc.Git.Type)
	}
	if mainRepo != "" && loc.Git.MainRepo != mainRepo {
		t.Fatalf("location[%d] git.main_repo: expected %s, got %s", idx, mainRepo, loc.Git.MainRepo)
	}
	if branch != "" && loc.Git.Branch != branch {
		t.Fatalf("location[%d] git.branch: expected %s, got %s", idx, branch, loc.Git.Branch)
	}
}

// --- git helpers ---

func skipIfNoGit(t *T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git not available")
	}
}

func runGit(t *T, dir string, args ...string) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func initGitRepo(t *T, path string) {
	runGit(t, path, "init")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# test")
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", "init")
}

// --- test functions ---

func testBasicMove(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	assertHistoryNil(t)
	runMvdOk(t, src, dst)
	assertHistoryChain(t, src, src, dst)
	assertFileExists(t, filepath.Join(dst, "f.txt"))
	assertFileNotExists(t, src)
}

func testMoveToExistingDir(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "mysrc")
	dst := filepath.Join(work, "existing-dir")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	runMvdOk(t, src, dst)
	finalPath := filepath.Join(dst, "mysrc")
	assertHistoryChain(t, src, src, finalPath)
	assertFileExists(t, filepath.Join(finalPath, "f.txt"))
}

func testMultiStepMove(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	d2 := filepath.Join(work, "d2")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))
	assertNoErr(t, os.MkdirAll(d2, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")
	assertFileExists(t, p1)

	runMvdOk(t, p1, d2)
	p2 := filepath.Join(d2, "src")
	assertFileExists(t, p2)

	assertHistoryChain(t, src, src, p1, p2)
}

func testMoveAsRootPath(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	d2 := filepath.Join(work, "d2")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))
	assertNoErr(t, os.MkdirAll(d2, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")

	// use original root path as source
	runMvdOk(t, src, d2)
	p2 := filepath.Join(d2, "src")

	assertHistoryChain(t, src, src, p1, p2)
	assertFileExists(t, p2)
	assertFileNotExists(t, p1)
}

func testMoveByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst := filepath.Join(work, "archive")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", src)
	assertHistoryChain(t, src, src)

	// from cwd, use basename shortcut
	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	runMvdOk(t, "kool", dst)
	p := filepath.Join(dst, "kool")
	assertHistoryChain(t, src, src, p)
	assertFileExists(t, p)
}

func testMoveByAlias(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst1 := filepath.Join(work, "scratch")
	dst2 := filepath.Join(work, "final")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst1, 0755))
	assertNoErr(t, os.MkdirAll(dst2, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, src, dst1)
	p1 := filepath.Join(dst1, "kool")

	runMvdOk(t, "--add-alias", "kk", "kool")

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	runMvdOk(t, "kk", dst2)
	p2 := filepath.Join(dst2, "kool")
	assertHistoryChain(t, src, src, p1, p2)
	assertFileExists(t, p2)
}

func testAmbiguousBasename(t *T) {
	work := testDir(t, "work")
	first := filepath.Join(work, "projects", "kool")
	second := filepath.Join(work, "projects", "v2", "kool")
	dst := filepath.Join(work, "dst")
	cwd := filepath.Join(work, "cwd")
	assertNoErr(t, os.MkdirAll(first, 0755))
	assertNoErr(t, os.MkdirAll(second, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	assertNoErr(t, os.MkdirAll(cwd, 0755))

	runMvdOk(t, "--add", first)
	runMvdOk(t, "--add", second)

	cwdOrig, _ := os.Getwd()
	assertNoErr(t, os.Chdir(cwd))
	defer os.Chdir(cwdOrig)

	out := runMvdErr(t, "kool", dst)
	assertContains(t, out, "ambiguous root basename kool")
}

func testAdd(t *T) {
	work := testDir(t, "work")
	dir := filepath.Join(work, "tracked")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	assertHistoryNil(t)
	runMvdOk(t, "--add", dir)
	assertHistoryChain(t, dir, dir)
}

func testAddDuplicate(t *T) {
	work := testDir(t, "work")
	dir := filepath.Join(work, "tracked")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	runMvdOk(t, "--add", dir)
	out := runMvdOk(t, "--add", dir)
	assertContains(t, out, "already recorded")
	assertHistoryLen(t, 1)
}

func testRemove(t *T) {
	work := testDir(t, "work")
	dir := filepath.Join(work, "tracked")
	assertNoErr(t, os.MkdirAll(dir, 0755))

	runMvdOk(t, "--add", dir)
	runMvdOk(t, "--rm", dir)
	assertHistoryNil(t)
}

func testRemoveForce(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))

	runMvdOk(t, src, d1)

	out := runMvdOk(t, "--rm", "-f", src)
	assertContains(t, out, "will clear")
	assertHistoryNil(t)
}

func testRemoveNoForceWithHistory(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))

	runMvdOk(t, src, d1)

	out := runMvdErr(t, "--rm", src)
	assertContains(t, out, "has movement history")
	assertHistoryContainsKey(t, src)
}

func testRebase(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	newBase := filepath.Join(work, "rebased")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")

	runMvdOk(t, "--rebase", src, newBase)
	assertHistoryChain(t, newBase, newBase, src, p1)
}

func testBack(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))
	writeFile(t, filepath.Join(src, "f.txt"), "hello")

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "src")

	runMvdOk(t, "--back", p)
	assertHistoryChain(t, src, src)
	assertFileExists(t, filepath.Join(src, "f.txt"))
	assertFileNotExists(t, p)
}

func testBackAtOrigin(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "src")

	// back
	runMvdOk(t, "--back", p)
	// back at origin is no-op
	out := runMvdOk(t, "--back", src)
	assertContains(t, out, "nothing to move back")
	assertHistoryChain(t, src, src)
}

func testBackByBasename(t *T) {
	work := testDir(t, "work")
	projectRoot := filepath.Join(work, "projects")
	src := filepath.Join(projectRoot, "kool")
	dst := filepath.Join(work, "scratch")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "kool")

	out := runMvdOk(t, "--back", "kool")
	assertContains(t, out, "moved back")
	assertHistoryChain(t, src, src)
	assertFileExists(t, src)
	assertFileNotExists(t, p)
}

func testListAll(t *T) {
	work := testDir(t, "work")
	s1 := filepath.Join(work, "proj1")
	s2 := filepath.Join(work, "proj2")
	assertNoErr(t, os.MkdirAll(s1, 0755))
	assertNoErr(t, os.MkdirAll(s2, 0755))

	runMvdOk(t, "--add", s1)
	runMvdOk(t, "--add", s2)

	out := runMvdOk(t, "--list")
	assertContains(t, out, s1)
	assertContains(t, out, s2)
}

func testListSingle(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	p := filepath.Join(dst, "src")
	assertFileExists(t, p)

	out := runMvdOk(t, "--list", src)
	assertContains(t, out, "(original)")
	assertContains(t, out, "*")
}

func testClear(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	dst := filepath.Join(work, "dst")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(dst, 0755))

	runMvdOk(t, src, dst)
	out := runMvdOk(t, "--clear", src)
	assertContains(t, out, "cleared history")
	assertHistoryNil(t)
}

func testMultiStepBack(t *T) {
	work := testDir(t, "work")
	src := filepath.Join(work, "src")
	d1 := filepath.Join(work, "d1")
	d2 := filepath.Join(work, "d2")
	assertNoErr(t, os.MkdirAll(src, 0755))
	assertNoErr(t, os.MkdirAll(d1, 0755))
	assertNoErr(t, os.MkdirAll(d2, 0755))

	runMvdOk(t, src, d1)
	p1 := filepath.Join(d1, "src")

	runMvdOk(t, p1, d2)
	p2 := filepath.Join(d2, "src")

	assertHistoryChain(t, src, src, p1, p2)

	// back step 1
	runMvdOk(t, "--back", p2)
	assertHistoryChain(t, src, src, p1)
	assertFileExists(t, p1)

	// back step 2
	runMvdOk(t, "--back", p1)
	assertHistoryChain(t, src, src)
	assertFileExists(t, src)

	// back at origin → no-op
	out := runMvdOk(t, "--back", src)
	assertContains(t, out, "nothing to move back")
}

func testNonExistentSrc(t *T) {
	work := testDir(t, "work")
	nonexistent := filepath.Join(work, "no-such-dir")
	dst := filepath.Join(work, "dst")

	out := runMvdErr(t, nonexistent, dst)
	assertContains(t, out, "no such file or directory")
}

// --- git-based tests ---

func testWorktreeMove(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")

	out := runMvdOk(t, "-w", mainRepo, wtDir)
	assertContains(t, out, "worktree created:")
	assertContains(t, out, "[branch: feature]")
	assertFileExists(t, filepath.Join(wtDir, ".git"))

	assertHistoryChain(t, mainRepo, mainRepo, wtDir)
	assertHistoryWorktreeEntry(t, mainRepo, 1, mainRepo, "feature")
}

func testWorktreeNonGitSrc(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	nonGit := filepath.Join(work, "not-a-repo")
	assertNoErr(t, os.MkdirAll(nonGit, 0755))
	wtDir := filepath.Join(work, "feature")

	out := runMvdErr(t, "-w", nonGit, wtDir)
	assertContains(t, out, "not a git repository")
	assertHistoryNil(t)
}

func testWorktreeBackDirty(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	runMvdOk(t, "-w", mainRepo, wtDir)

	writeFile(t, filepath.Join(wtDir, "dirty-file"), "uncommitted")

	out := runMvdErr(t, "--back", wtDir)
	assertContains(t, out, "uncommitted changes")
	assertFileExists(t, wtDir)
}

func testWorktreeBackUnmerged(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	runMvdOk(t, "-w", mainRepo, wtDir)

	writeFile(t, filepath.Join(wtDir, "feature-work"), "work")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	out := runMvdErr(t, "--back", wtDir)
	assertContains(t, out, "not merged")
	assertFileExists(t, wtDir)
}

func testWorktreeBackSuccess(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	wtDir := filepath.Join(work, "feature")
	runMvdOk(t, "-w", mainRepo, wtDir)

	writeFile(t, filepath.Join(wtDir, "feature-work"), "work")
	runGit(t, wtDir, "add", "feature-work")
	runGit(t, wtDir, "commit", "-m", "feature work")

	runGit(t, mainRepo, "merge", "feature")

	out := runMvdOk(t, "--back", wtDir)
	assertContains(t, out, "worktree removed:")
	assertContains(t, out, "branch: feature deleted")
	assertFileNotExists(t, wtDir)
	assertHistoryNil(t)
}

func testWorktreeBranchCollision(t *T) {
	skipIfNoGit(t)
	work := testDir(t, "work")

	mainRepo := filepath.Join(work, "main")
	assertNoErr(t, os.MkdirAll(mainRepo, 0755))
	initGitRepo(t, mainRepo)

	// create a branch that matches the basename
	runGit(t, mainRepo, "branch", "myfeature")

	wtDir := filepath.Join(work, "myfeature")
	out := runMvdOk(t, "-w", mainRepo, wtDir)

	// branch should have date suffix, not plain "myfeature"
	assertContains(t, out, "worktree created:")
	assertNotContains(t, out, "[branch: myfeature]")

	assertFileExists(t, filepath.Join(wtDir, ".git"))
	assertHistoryChain(t, mainRepo, mainRepo, wtDir)

	// verify branch is not named exactly "myfeature"
	h := readHistory(t)
	proj := h.Projects[mainRepo]
	branch := proj.Locations[1].Git.Branch
	if branch == "myfeature" {
		t.Fatalf("branch name should not be 'myfeature' (collision)")
	}
	if !strings.HasPrefix(branch, "myfeature-") {
		t.Fatalf("branch name should start with 'myfeature-', got %q", branch)
	}
}
