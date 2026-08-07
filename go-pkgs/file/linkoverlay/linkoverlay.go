// Package linkoverlay merges directory and sparse-file layers into a target tree
// using absolute top-level symlinks and explode-on-demand intermediates.
package linkoverlay

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// File is a sparse overlay entry written under the merge target.
type File struct {
	Path    string // relative path under target
	Mode    uint32 // 0 → treat as 0o644
	Content []byte
}

// Layer is one merge layer: optional base Dir plus optional sparse Files.
// Within a layer, Dir is applied first, then Files.
type Layer struct {
	Dir   string // optional base directory
	Files []File // optional sparse overlay
}

// Apply merges layers into target left→right; later layers win.
// Within a layer: if both Dir and Files set, apply Dir first, then Files.
func Apply(target string, layers ...Layer) error {
	if target == "" {
		return fmt.Errorf("empty target")
	}
	if err := os.MkdirAll(target, 0o700); err != nil {
		return err
	}
	absTarget, err := filepath.Abs(target)
	if err != nil {
		return err
	}

	for _, layer := range layers {
		if layer.Dir != "" {
			if err := applyDir(absTarget, layer.Dir); err != nil {
				return err
			}
		}
		for _, f := range layer.Files {
			if err := applyFile(absTarget, f); err != nil {
				return err
			}
		}
	}
	return nil
}

// ApplyDirs is Apply(target, Layer{Dir: d0}, Layer{Dir: d1}, ...).
func ApplyDirs(target string, dirs ...string) error {
	layers := make([]Layer, 0, len(dirs))
	for _, d := range dirs {
		layers = append(layers, Layer{Dir: d})
	}
	return Apply(target, layers...)
}

// applyDir projects each top-level name under dir as an absolute symlink into
// target (including dot entries; skip . and ..). Existing target entries with
// the same name are replaced so later Dir layers win.
func applyDir(target, dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	ents, err := os.ReadDir(absDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range ents {
		name := e.Name()
		if name == "." || name == ".." {
			continue
		}
		absEntry, err := filepath.Abs(filepath.Join(absDir, name))
		if err != nil {
			return err
		}
		linkPath := filepath.Join(target, name)
		// Later Dir wins over prior seed symlinks or exploded dirs.
		if err := os.RemoveAll(linkPath); err != nil {
			return err
		}
		if err := os.Symlink(absEntry, linkPath); err != nil {
			return fmt.Errorf("seed symlink %s: %w", name, err)
		}
	}
	return nil
}

// applyFile writes one sparse overlay file under target, exploding intermediate
// symlinks and replacing the leaf without following into a base.
func applyFile(target string, f File) error {
	rel, err := sanitizeRelPath(f.Path)
	if err != nil {
		return err
	}
	dest := filepath.Join(target, rel)
	if err := ensureUnderRoot(target, dest, f.Path); err != nil {
		return err
	}

	slashRel := filepath.ToSlash(rel)
	parts := strings.Split(slashRel, "/")
	cur := target
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]
		if part == "" || part == "." {
			continue
		}
		cur = filepath.Join(cur, part)
		if err := ensureDirExploded(cur); err != nil {
			return fmt.Errorf("explode %s: %w", filepath.ToSlash(strings.Join(parts[:i+1], "/")), err)
		}
	}

	// Remove first so WriteFile does not follow a seed symlink into the base.
	_ = os.Remove(dest)
	return writeOverlayFile(dest, f)
}

// ensureDirExploded makes path a real directory suitable for writing children.
// If path is a symlink, it is unlinked, replaced with a directory, and children
// of the former link target are re-linked with absolute targets.
func ensureDirExploded(path string) error {
	st, err := os.Lstat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(path, 0o700)
		}
		return err
	}

	if st.Mode()&os.ModeSymlink != 0 {
		linkTarget, err := os.Readlink(path)
		if err != nil {
			return err
		}
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Join(filepath.Dir(path), linkTarget)
		}
		absReal, err := filepath.Abs(linkTarget)
		if err != nil {
			return err
		}
		if err := os.Remove(path); err != nil {
			return err
		}
		if err := os.Mkdir(path, 0o700); err != nil {
			return err
		}
		return reLinkChildren(path, absReal)
	}

	if st.IsDir() {
		return nil
	}

	// Wrong type (regular file, etc.): replace with empty directory.
	if err := os.RemoveAll(path); err != nil {
		return err
	}
	return os.Mkdir(path, 0o700)
}

// reLinkChildren creates absolute symlinks in sandboxDir for each entry under realDir.
func reLinkChildren(sandboxDir, realDir string) error {
	ents, err := os.ReadDir(realDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range ents {
		name := e.Name()
		if name == "." || name == ".." {
			continue
		}
		absTarget, err := filepath.Abs(filepath.Join(realDir, name))
		if err != nil {
			return err
		}
		linkPath := filepath.Join(sandboxDir, name)
		if err := os.Symlink(absTarget, linkPath); err != nil {
			return err
		}
	}
	return nil
}

func writeOverlayFile(dest string, f File) error {
	mode := os.FileMode(f.Mode) & os.ModePerm
	if mode == 0 {
		mode = 0o644
	}
	if err := os.WriteFile(dest, f.Content, mode); err != nil {
		return err
	}
	return os.Chmod(dest, mode)
}

func sanitizeRelPath(path string) (string, error) {
	rel := filepath.Clean(filepath.FromSlash(path))
	if rel == "." || rel == "" {
		return "", fmt.Errorf("invalid empty path")
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path: %s", path)
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("invalid absolute path: %s", path)
	}
	return rel, nil
}

func ensureUnderRoot(root, dest, orig string) error {
	absDest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}
	relCheck, err := filepath.Rel(absRoot, absDest)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(filepath.Separator)) {
		return fmt.Errorf("path escapes sandbox root: %s", orig)
	}
	return nil
}
