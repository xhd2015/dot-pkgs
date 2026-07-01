// Package scan walks a directory tree for Go modules, applying skip rules
// (.git/vendor/testdata, gitignored dirs, and nested separate git repos) and
// returns or streams the modules found.
//
// Scan collects all modules and sorts them by Dir (lexical); ScanStream calls a
// function for each module in walk order as it is discovered (no sort, no
// buffering).
package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/xhd2015/gitops/git"
	"golang.org/x/mod/modfile"
)

// Module describes a single Go module discovered under the scan root.
//
// Dir is the module's directory relative to the scan root: "." for the root,
// slash-joined (e.g. "app", "nested/service") for sub-directories, with no "./"
// prefix. Path is the module path from go.mod. Requires and Replaces mirror the
// go.mod require/replace blocks.
type Module struct {
	Dir      string
	Path     string
	Requires []ModuleRequire
	Replaces []ModuleReplace
}

// ModuleRequire is a single require entry from go.mod.
type ModuleRequire struct {
	Path    string
	Version string
}

// ModuleReplace is a single replace entry from go.mod.
type ModuleReplace struct {
	OldPath    string
	NewPath    string
	NewVersion string
}

// HasLocalFilesystemReplace reports whether m contains a filesystem replace
// directive — a replace whose target is a relative (./ or ../) or absolute path
// with no version. Such replaces point at a local checkout rather than a
// published module version. Mirrors resolve.HasLocalFilesystemReplace but
// operates on a scanned Module.
func (m Module) HasLocalFilesystemReplace() bool {
	for _, repl := range m.Replaces {
		p := repl.NewPath
		if p == "" || repl.NewVersion != "" {
			continue
		}
		if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || filepath.IsAbs(p) {
			return true
		}
	}
	return false
}

// LocalFilesystemReplaces returns the replace directives of m whose target is a
// filesystem/local path — a relative (./ or ../) or absolute path with no
// version. Such replaces point at a local checkout rather than a published
// module version. It is the returning form of HasLocalFilesystemReplace so
// callers can render the offending directives.
func (m Module) LocalFilesystemReplaces() []ModuleReplace {
	var out []ModuleReplace
	for _, repl := range m.Replaces {
		p := repl.NewPath
		if p == "" || repl.NewVersion != "" {
			continue
		}
		if strings.HasPrefix(p, "./") || strings.HasPrefix(p, "../") || filepath.IsAbs(p) {
			out = append(out, repl)
		}
	}
	return out
}

// Options configures Scan / ScanStream. The zero value applies all skip rules
// (name-based, gitignored, nested-separate-repo). When the root is not a git
// repo the git-based skips are disabled automatically.
type Options struct {
	// Reserved for future toggles.
}

// Scan walks root and returns all Go modules found, sorted by Dir (lexical).
// It is implemented in terms of ScanStream: collect, then sort.
func Scan(root string, opts Options) ([]Module, error) {
	var modules []Module
	err := ScanStream(root, opts, func(m Module) error {
		modules = append(modules, m)
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Dir < modules[j].Dir
	})
	return modules, nil
}

// ScanStream walks root and calls fn for each Go module in walk order as it is
// found. No sorting is performed. If fn returns an error, walking stops and
// that error is returned.
func ScanStream(root string, opts Options, fn func(Module) error) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	skipper, err := newSkipper(absRoot)
	if err != nil {
		return err
	}

	return filepath.WalkDir(absRoot, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !d.IsDir() {
			return nil
		}

		if path != absRoot {
			skip, err := skipper.shouldSkip(absRoot, path)
			if err != nil {
				return err
			}
			if skip {
				return filepath.SkipDir
			}
		}

		ok, err := hasGoMod(path)
		if err != nil {
			return err
		}
		if !ok {
			return nil
		}

		rel, err := filepath.Rel(absRoot, path)
		if err != nil {
			return err
		}
		module, err := readModule(path, filepath.ToSlash(rel))
		if err != nil {
			return err
		}
		return fn(module)
	})
}

// skipper encapsulates the git-based skip rules. When the root is not a git
// repo, only name-based skips apply and the skipper is effectively inert.
type skipper struct {
	inGit       bool
	root        string
	ignoredDirs map[string]struct{}
}

func newSkipper(absRoot string) (*skipper, error) {
	s := &skipper{root: absRoot}
	inside, err := git.IsInsideGit(absRoot)
	if err != nil {
		return nil, err
	}
	if !inside {
		return s, nil
	}
	s.inGit = true

	ignored, err := git.ListIgnoredDirs(absRoot)
	if err != nil {
		return nil, err
	}
	s.ignoredDirs = make(map[string]struct{}, len(ignored))
	for _, dir := range ignored {
		s.ignoredDirs[dir] = struct{}{}
	}
	return s, nil
}

// shouldSkip reports whether a non-root directory should be pruned, applying the
// skip rules in order: name-based, gitignore, nested-separate-repo.
func (s *skipper) shouldSkip(absRoot, path string) (bool, error) {
	name := filepath.Base(path)
	switch name {
	case ".git", "vendor", "testdata":
		return true, nil
	}

	if !s.inGit {
		return false, nil
	}

	rel, err := filepath.Rel(absRoot, path)
	if err != nil {
		return false, err
	}
	relSlash := filepath.ToSlash(filepath.Clean(rel))

	// gitignored (root's .gitignore / exclude rules, including nested .gitignore
	// files that git honors). Check the ignored-dir batch first, then fall back to
	// a per-path check-ignore so nested .gitignore patterns are respected.
	if s.isIgnoredDir(relSlash) {
		return true, nil
	}
	ignored, err := git.CheckIgnore(absRoot, relSlash)
	if err != nil {
		return false, err
	}
	if ignored {
		return true, nil
	}

	// Nested separate repo: a directory that carries its own .git but is NOT a
	// tracked submodule/gitlink of the root repo. Its whole subtree is skipped.
	if hasOwnGit(path) {
		isSub, err := git.IsSubmodule(absRoot, relSlash)
		if err != nil {
			return false, err
		}
		if !isSub {
			return true, nil
		}
	}

	return false, nil
}

// isIgnoredDir reports whether relSlash or any of its ancestor directories is in
// the upfront ignored-dir batch (so an ignored parent prunes its children).
func (s *skipper) isIgnoredDir(relSlash string) bool {
	for current := relSlash; current != "." && current != ""; {
		if _, ok := s.ignoredDirs[current]; ok {
			return true
		}
		next := filepath.Dir(current)
		if next == current {
			break
		}
		current = next
	}
	return false
}

func hasOwnGit(dir string) bool {
	info, err := os.Stat(filepath.Join(dir, ".git"))
	if err != nil {
		return false
	}
	_ = info
	return true
}

func hasGoMod(dir string) (bool, error) {
	info, err := os.Stat(filepath.Join(dir, "go.mod"))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return !info.IsDir(), nil
}

func readModule(dir, rel string) (Module, error) {
	goModPath := filepath.Join(dir, "go.mod")
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return Module{}, err
	}

	module := Module{
		Dir:  rel,
		Path: modfile.ModulePath(data),
	}

	modFile, err := modfile.Parse(goModPath, data, nil)
	if err != nil {
		// Fall back to lax parsing for go.mod files with template placeholders
		// or other non-standard content (e.g. "module example.com/{{.Name}}").
		modFile, err = modfile.ParseLax(goModPath, data, nil)
		if err != nil {
			return module, nil
		}
	}
	if modFile.Module != nil && modFile.Module.Mod.Path != "" {
		module.Path = modFile.Module.Mod.Path
	}

	for _, req := range modFile.Require {
		module.Requires = append(module.Requires, ModuleRequire{
			Path:    req.Mod.Path,
			Version: req.Mod.Version,
		})
	}
	for _, repl := range modFile.Replace {
		module.Replaces = append(module.Replaces, ModuleReplace{
			OldPath:    repl.Old.Path,
			NewPath:    repl.New.Path,
			NewVersion: repl.New.Version,
		})
	}

	return module, nil
}
