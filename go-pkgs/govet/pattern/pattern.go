package pattern

import (
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

func ResolvePatterns(workDir string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		patterns = []string{"."}
	}

	var wildcardPatterns []string
	var plainDirs []string

	for _, p := range patterns {
		if strings.Contains(p, "...") {
			wildcardPatterns = append(wildcardPatterns, p)
		} else {
			plainDirs = append(plainDirs, p)
		}
	}

	dirSet := make(map[string]bool)
	for _, d := range plainDirs {
		if !filepath.IsAbs(d) {
			d = filepath.Join(workDir, d)
		}
		dirSet[d] = true
	}

	if len(wildcardPatterns) > 0 {
		cfg := &packages.Config{
			Mode:  packages.NeedFiles,
			Dir:   workDir,
			Tests: true,
		}

		pkgs, err := packages.Load(cfg, wildcardPatterns...)
		if err != nil {
			return nil, err
		}

		for _, pkg := range pkgs {
			if pkg.Dir == "" {
				continue
			}
			if len(pkg.Errors) > 0 && len(pkg.GoFiles) == 0 {
				continue
			}
			dirSet[pkg.Dir] = true
		}
	}

	var dirs []string
	for dir := range dirSet {
		dirs = append(dirs, dir)
	}
	sort.Strings(dirs)
	return dirs, nil
}
