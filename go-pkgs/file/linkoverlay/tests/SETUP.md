# Scenario

**Feature**: `linkoverlay.Apply` / `ApplyDirs` merge Dir seeds and Files overlays into a target

```
# workspace per leaf under t.TempDir; no process env/cwd mutation
leaf Setup -> materialize base dirs + LayerSpecs
caller target + layers -> Apply | ApplyDirs -> target tree (abs symlinks / files)
Assert <- Lstat/Readlink/ReadFile under target (+ base unchanged checks)
```

## Preconditions

- Package `github.com/xhd2015/dot-pkgs/go-pkgs/file/linkoverlay` is importable
  (scaffold stub allowed; real merge lands in implementer phase).
- All leaves are L2 in-process: call package API directly; inspect filesystem.
- Parallel-safe: each leaf uses `t.TempDir()` via root Setup; no `os.Setenv` /
  `t.Setenv` / `os.Chdir` / `t.Chdir`.
- Platform: macOS + Linux (symlinks required).

## Steps

1. Root `Setup` creates `req.WorkingDir` (`t.TempDir`) and ensures empty
   `target/` under it; zeros request fields.
2. Leaf/grouping `Setup` sets `UseApplyDirs` / `Layers` / `DirsRel` and may call
   `materializeLayers` to write base fixture trees.
3. Root `Run` calls `ApplyDirs` or `Apply` once and returns absolute target.
4. Leaf `Assert` checks error and target tree (and base immutability when needed).

## Context

- **Dir projection:** top-level names only (including dots); absolute symlink targets.
- **Explode:** intermediate seed symlink → real dir + re-link children from the
  symlink’s former target directory.
- **Leaf write:** remove existing (no follow into base) then write content + mode.
- Path safety errors must mention invalid/absolute/`..` — stub `"not implemented"`
  alone is not a valid path-reject outcome (keeps those leaves RED on stub).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	_ = d
	req.WorkingDir = t.TempDir()
	req.TargetRel = "target"
	req.UseApplyDirs = false
	req.DirsRel = nil
	req.Layers = nil

	target := filepath.Join(req.WorkingDir, req.TargetRel)
	if err := os.MkdirAll(target, 0o755); err != nil {
		return err
	}
	return nil
}

// materializeLayers writes each LayerSpec.BaseFiles under WorkingDir/DirRel.
// Call after leaves assign req.Layers. Idempotent for empty BaseFiles (still MkdirAll).
func materializeLayers(t *testing.T, req *Request) {
	t.Helper()
	for _, layer := range req.Layers {
		if layer.DirRel == "" {
			continue
		}
		base := filepath.Join(req.WorkingDir, layer.DirRel)
		if err := os.MkdirAll(base, 0o755); err != nil {
			t.Fatal(err)
		}
		writeTree(t, base, layer.BaseFiles)
	}
}

func writeTree(t *testing.T, root string, files []FileSpec) {
	t.Helper()
	for _, f := range files {
		rel := filepath.FromSlash(f.Path)
		dest := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			t.Fatal(err)
		}
		mode := os.FileMode(f.Mode) & os.ModePerm
		if mode == 0 {
			mode = 0o644
		}
		if err := os.WriteFile(dest, []byte(f.Content), mode); err != nil {
			t.Fatal(err)
		}
	}
}

func absJoin(parts ...string) string {
	return filepath.Clean(filepath.Join(parts...))
}

func mustLstat(t *testing.T, path string) os.FileInfo {
	t.Helper()
	st, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("Lstat %s: %v", path, err)
	}
	return st
}

func assertSymlinkTo(t *testing.T, link, wantTarget string) {
	t.Helper()
	st := mustLstat(t, link)
	if st.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("%s: want symlink, got mode %v", link, st.Mode())
	}
	got, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("Readlink %s: %v", link, err)
	}
	gotAbs, err := filepath.Abs(got)
	if err != nil {
		t.Fatalf("Abs link target %s: %v", got, err)
	}
	wantAbs, err := filepath.Abs(wantTarget)
	if err != nil {
		t.Fatalf("Abs want %s: %v", wantTarget, err)
	}
	if gotAbs != wantAbs {
		t.Fatalf("symlink %s -> %q want %q", link, gotAbs, wantAbs)
	}
}

func assertRegularContent(t *testing.T, path, want string) {
	t.Helper()
	st := mustLstat(t, path)
	if st.Mode()&os.ModeSymlink != 0 {
		// Following is OK for content checks when leaf may still be a symlink
		// into a later base; use ReadFile which follows.
	} else if !st.Mode().IsRegular() && st.Mode()&os.ModeType != 0 {
		// allow regular only when not symlink; directories fail here
		if !st.IsDir() {
			// non-regular non-symlink
			t.Fatalf("%s: unexpected mode %v", path, st.Mode())
		}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("%s content = %q want %q", path, string(data), want)
	}
}

func assertFileContentUnchanged(t *testing.T, path, want string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile base %s: %v", path, err)
	}
	if string(data) != want {
		t.Fatalf("base %s mutated: got %q want %q", path, string(data), want)
	}
}

func isNotImplemented(err error) bool {
	return err != nil && strings.Contains(err.Error(), "not implemented")
}

func assertPathSafetyError(t *testing.T, err error, hints ...string) {
	t.Helper()
	if err == nil {
		t.Fatal("expected path safety error, got nil")
	}
	if isNotImplemented(err) {
		t.Fatalf("stub error is not path validation: %v", err)
	}
	msg := strings.ToLower(err.Error())
	for _, h := range hints {
		if strings.Contains(msg, strings.ToLower(h)) {
			return
		}
	}
	// Also accept generic "invalid" / "escape" / "absolute" language.
	for _, soft := range []string{"invalid", "escape", "absolute", ".."} {
		if strings.Contains(msg, soft) {
			return
		}
	}
	t.Fatalf("error %q does not look like path safety (hints=%v)", err, hints)
}

func targetPath(req *Request, resp *Response, rel string) string {
	base := resp.Target
	if base == "" {
		tr := req.TargetRel
		if tr == "" {
			tr = "target"
		}
		base = filepath.Join(req.WorkingDir, tr)
	}
	return filepath.Join(base, filepath.FromSlash(rel))
}

func basePath(req *Request, dirRel, fileRel string) string {
	return filepath.Join(req.WorkingDir, dirRel, filepath.FromSlash(fileRel))
}
```
