package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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
		{"TestMoveNonExistentBasename", testMoveNonExistentBasename},
		{"TestWorktreeMove", testWorktreeMove},
		{"TestWorktreeNonGitSrc", testWorktreeNonGitSrc},
		{"TestWorktreeBackDirty", testWorktreeBackDirty},
		{"TestWorktreeBackUnmerged", testWorktreeBackUnmerged},
		{"TestWorktreeBackUnmergedConfirm", testWorktreeBackUnmergedConfirm},
		{"TestWorktreeBackSuccess", testWorktreeBackSuccess},
		{"TestWorktreeBranchCollision", testWorktreeBranchCollision},
		{"TestMoveWorktreeWithoutWFlagShouldDoSimpleMove", testMoveWorktreeWithoutWFlagShouldDoSimpleMove},
		{"TestMoveNestedWorktreeWithoutWFlag", testMoveNestedWorktreeWithoutWFlag},
		{"TestMoveWorktreeWithWFlagShouldRunGitWorktreeAdd", testMoveWorktreeWithWFlagShouldRunGitWorktreeAdd},
		{"TestWorktreeMoveByBasename", testWorktreeMoveByBasename},
		{"TestClearByBasename", testClearByBasename},
		{"TestRebaseByBasename", testRebaseByBasename},
		{"TestListByBasename", testListByBasename},
		{"TestClearWithDollarExpansion", testClearWithDollarExpansion},
		{"TestBackWithDollarExpansion", testBackWithDollarExpansion},
		{"TestListWithDollarExpansion", testListWithDollarExpansion},
		{"TestRebaseWithDollarExpansion", testRebaseWithDollarExpansion},
		{"TestWorktreeMoveWithDollarExpansion", testWorktreeMoveWithDollarExpansion},
		{"TestWhichWithDollarExpansion", testWhichWithDollarExpansion},
		{"TestMoveDefaultWithDollarExpansion", testMoveDefaultWithDollarExpansion},
		{"TestAddWithDollarExpansion", testAddWithDollarExpansion},
		{"TestAddNonExistentFails", testAddNonExistentFails},
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

func buildMvd() (string, error) {
	modRoot := findModuleRoot()
	if modRoot == "" {
		return "", fmt.Errorf("cannot find module root (no go.mod found)")
	}
	bin := filepath.Join(os.TempDir(), "mvd-integ-test")
	mvdPkg := "."
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
	t.Logf("%s", string(out))
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
