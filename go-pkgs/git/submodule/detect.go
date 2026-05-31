package submodule

import (
	"os"
	"path/filepath"
)

func DetectSubModules(files []string) []string {
	seen := make(map[string]bool)
	var subModules []string
	for _, f := range files {
		checkDir := func(dir string) bool {
			gitPath := filepath.Join(dir, ".git")
			info, err := os.Stat(gitPath)
			if err == nil && (info.IsDir() || info.Mode().IsRegular()) {
				if !seen[dir] {
					seen[dir] = true
					subModules = append(subModules, dir)
				}
				return true
			}
			return false
		}

		if checkDir(f) {
			continue
		}

		dir := filepath.Dir(f)
		for {
			if dir == "." {
				break
			}
			if checkDir(dir) {
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return subModules
}
