## Preconditions
- The mvd Go module is located two levels above the test tree root (at `go-pkgs/cmd/`)
- Go toolchain is available on PATH
- Git may be available for worktree tests

## Context
Each test runs the mvd CLI tool in an isolated environment. The binary is built once and reused across all tests. Each leaf gets its own temp directory and isolated `MVD_DEBUG_CONFIG_HOME`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

var buildOnce sync.Once
var builtMvdBin string
var buildMvdErr error

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

func getMvdBin(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		modRoot := filepath.Dir(filepath.Dir(DOCTEST_ROOT))
		if modRoot == "" {
			modRoot = findModuleRoot(DOCTEST_ROOT)
		}
		tmpDir, err := os.MkdirTemp("", "mvd-doc-test")
		if err != nil {
			buildMvdErr = fmt.Errorf("create temp dir: %w", err)
			return
		}
		bin := filepath.Join(tmpDir, "mvd")
		cmd := exec.Command("go", "build", "-o", bin, "./mvd")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			buildMvdErr = fmt.Errorf("build mvd: %w\n%s", err, out)
			return
		}
		builtMvdBin = bin
	})
	if buildMvdErr != nil {
		t.Fatal(buildMvdErr)
	}
	return builtMvdBin
}

type GitInfo struct {
	Type     string `json:"type"`
	MainRepo string `json:"main_repo,omitempty"`
	Branch   string `json:"branch,omitempty"`
}

type LocationEntry struct {
	Path string   `json:"path"`
	Git  *GitInfo `json:"git,omitempty"`
}

type MoveEntry struct {
	Prev    string `json:"prev"`
	Current string `json:"current"`
	Type    string `json:"type"`
	Branch  string `json:"branch,omitempty"`
}

type ProjectEntry struct {
	Root      string          `json:"root,omitempty"`
	Locations []LocationEntry `json:"locations,omitempty"`
	Moves     []MoveEntry     `json:"moves,omitempty"`
	Aliases   []string        `json:"aliases,omitempty"`
}

type HistoryFile struct {
	Version  string                  `json:"version"`
	Projects map[string]ProjectEntry `json:"projects"`
}

type Request struct {
	ConfigHome string
	WorkRoot   string
	Args       []string
}

type Response struct {
	Output   string
	ExitCode int
}

func runMvd(t *testing.T, req *Request) (*Response, error) {
	bin := getMvdBin(t)
	cmd := exec.Command(bin, req.Args...)
	cmd.Env = append(os.Environ(), "MVD_DEBUG_CONFIG_HOME="+req.ConfigHome)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{Output: string(out), ExitCode: exitCode}, nil
}

func Setup(t *testing.T, req *Request) error {
	req.WorkRoot = t.TempDir()
	req.ConfigHome = filepath.Join(req.WorkRoot, ".mvd-config")
	ensureHelpersUsed()
	return os.MkdirAll(req.ConfigHome, 0755)
}

func Run(t *testing.T, req *Request) (*Response, error) {
	bin := getMvdBin(t)
	cmd := exec.Command(bin, req.Args...)
	cmd.Env = append(os.Environ(), "MVD_DEBUG_CONFIG_HOME="+req.ConfigHome)
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, err
		}
	}
	return &Response{Output: string(out), ExitCode: exitCode}, nil
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

func writeHistoryFile(t *testing.T, configHome string, hf HistoryFile) {
	t.Helper()
	data, err := json.Marshal(hf)
	if err != nil {
		t.Fatalf("marshal history: %v", err)
	}
	writeFile(t, filepath.Join(configHome, "history.json"), string(data))
}

func writeAliasesFile(t *testing.T, configHome string, aliases map[string]string) {
	t.Helper()
	data, err := json.Marshal(aliases)
	if err != nil {
		t.Fatalf("marshal aliases: %v", err)
	}
	writeFile(t, filepath.Join(configHome, "aliases.json"), string(data))
}

func locationsFromMoves(root string, moves []MoveEntry) []LocationEntry {
	locs := []LocationEntry{{Path: root}}
	for _, move := range moves {
		loc := LocationEntry{Path: move.Current}
		if move.Type == "worktree" {
			loc.Git = &GitInfo{
				Type:     "worktree",
				MainRepo: move.Prev,
				Branch:   move.Branch,
			}
		}
		locs = append(locs, loc)
	}
	return locs
}

func readHistoryFile(t *testing.T, configHome string) *HistoryFile {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(configHome, "history.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatalf("read history: %v", err)
	}
	var hf HistoryFile
	if err := json.Unmarshal(data, &hf); err != nil {
		t.Fatalf("parse history: %v", err)
	}
	if hf.Projects != nil {
		for key, proj := range hf.Projects {
			if len(proj.Moves) > 0 && len(proj.Locations) == 0 {
				proj.Locations = locationsFromMoves(proj.Root, proj.Moves)
				hf.Projects[key] = proj
			} else if len(proj.Locations) == 0 && proj.Root != "" {
				proj.Locations = []LocationEntry{{Path: proj.Root}}
				hf.Projects[key] = proj
			}
		}
	}
	return &hf
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("%s should exist", path)
	}
}

func assertFileNotExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("%s should not exist", path)
	}
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q in output, got:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Fatalf("expected %q not in output, got:\n%s", substr, s)
	}
}

func assertHistoryNil(t *testing.T, configHome string) {
	t.Helper()
	h := readHistoryFile(t, configHome)
	if h != nil && len(h.Projects) > 0 {
		t.Fatalf("expected no history, got %d projects", len(h.Projects))
	}
}

func assertHistoryLen(t *testing.T, configHome string, n int) *HistoryFile {
	t.Helper()
	h := readHistoryFile(t, configHome)
	if h == nil {
		t.Fatalf("expected history, got nil")
	}
	if len(h.Projects) != n {
		t.Fatalf("expected %d projects, got %d", n, len(h.Projects))
	}
	return h
}

func assertHistoryChain(t *testing.T, configHome string, key string, wantPaths ...string) {
	t.Helper()
	h := assertHistoryLen(t, configHome, 1)
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
			t.Fatalf("project key %s not found in history", key)
		}
	}
	if len(proj.Locations) != len(wantPaths) {
		t.Fatalf("expected %d locations, got %d", len(wantPaths), len(proj.Locations))
	}
	for i, want := range wantPaths {
		if proj.Locations[i].Path != want {
			t.Fatalf("location[%d]: expected %s, got %s", i, want, proj.Locations[i].Path)
		}
	}
}

func assertHistoryWorktreeEntry(t *testing.T, configHome string, key string, idx int, mainRepo, branch string) {
	t.Helper()
	h := readHistoryFile(t, configHome)
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
			t.Fatalf("project key %s not found in history", key)
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

func initGitRepo(t *testing.T, path string) {
	t.Helper()
	runGit(t, path, "init")
	runGit(t, path, "config", "user.email", "test@test.com")
	runGit(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# test")
	runGit(t, path, "add", "README.md")
	runGit(t, path, "commit", "-m", "init")
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Called from Setup to prevent Go's "declared and not used" errors.
// The generator places all helper functions inside a single function
// body where Go enforces usage. Referencing each helper here ensures
// they count as "used" even if a particular test leaf skips them.
func ensureHelpersUsed() {
	_ = mkdirAll
	_ = writeFile
	_ = writeHistoryFile
	_ = writeAliasesFile
	_ = readHistoryFile
	_ = assertFileExists
	_ = assertFileNotExists
	_ = assertContains
	_ = assertNotContains
	_ = assertHistoryNil
	_ = assertHistoryLen
	_ = assertHistoryChain
	_ = assertHistoryWorktreeEntry
	_ = skipIfNoGit
	_ = runGit
	_ = initGitRepo
	_ = assertErrIsNil
	_ = locationsFromMoves
	_ = runMvd
}
```
