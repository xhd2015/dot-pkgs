package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func ResolveGoFiles(workDir string, paths []string) ([]string, error) {
	if len(paths) == 0 {
		return nil, nil
	}

	var wildcardPatterns []string
	var plainPaths []string

	for _, p := range paths {
		if strings.Contains(p, "...") {
			wildcardPatterns = append(wildcardPatterns, p)
		} else {
			plainPaths = append(plainPaths, p)
		}
	}

	fileSet := make(map[string]bool)

	for _, p := range wildcardPatterns {
		files, err := resolveWildcard(workDir, p)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			fileSet[f] = true
		}
	}

	for _, p := range plainPaths {
		abs, err := resolvePath(workDir, p)
		if err != nil {
			return nil, err
		}
		files, err := collectGoFiles(abs)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			fileSet[f] = true
		}
	}

	var result []string
	for f := range fileSet {
		result = append(result, f)
	}
	sort.Strings(result)
	return result, nil
}

func resolvePath(workDir, p string) (string, error) {
	if filepath.IsAbs(p) {
		return p, nil
	}
	return filepath.Join(workDir, p), nil
}

func resolveWildcard(workDir string, pattern string) ([]string, error) {
	cfg := &packages.Config{
		Mode:  packages.NeedFiles | packages.NeedName,
		Dir:   workDir,
		Tests: true,
	}

	pkgs, err := packages.Load(cfg, pattern)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, pkg := range pkgs {
		if pkg.Dir == "" {
			continue
		}
		for _, f := range pkg.GoFiles {
			if !filepath.IsAbs(f) {
				f = filepath.Join(workDir, f)
			}
			files = append(files, f)
		}
	}
	return files, nil
}

func collectGoFiles(path string) ([]string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return walkGoFiles(path)
	}

	if strings.HasSuffix(path, ".go") {
		if isVendorFile(path) {
			return nil, nil
		}
		return []string{path}, nil
	}

	return nil, nil
}

func walkGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			base := filepath.Base(path)
			if base == ".git" {
				return filepath.SkipDir
			}
			if base == "vendor" && hasGoMod(filepath.Dir(path)) {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func hasGoMod(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, "go.mod"))
	return err == nil
}

func isVendorFile(path string) bool {
	dir := filepath.Dir(path)
	for dir != "." && dir != "/" {
		if filepath.Base(dir) == "vendor" && hasGoMod(filepath.Dir(dir)) {
			return true
		}
		dir = filepath.Dir(dir)
	}
	return false
}
