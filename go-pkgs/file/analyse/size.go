package analyse

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/detect"
)

// FormatSize renders a byte count as a human-readable string.
func FormatSize(n int64) string {
	const (
		kb = 1024
		mb = kb * 1024
		gb = mb * 1024
	)
	switch {
	case n >= gb:
		return fmt.Sprintf("%.2f GB", float64(n)/float64(gb))
	case n >= mb:
		return fmt.Sprintf("%.2f MB", float64(n)/float64(mb))
	case n >= kb:
		return fmt.Sprintf("%.2f KB", float64(n)/float64(kb))
	default:
		return fmt.Sprintf("%d B", n)
	}
}

func deepSize(root string) (int64, error) {
	info, err := os.Stat(root)
	if err != nil {
		return 0, err
	}
	if !info.IsDir() {
		return info.Size(), nil
	}

	var total int64
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsPermission(walkErr) {
				return nil
			}
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		fi, statErr := d.Info()
		if statErr != nil {
			if os.IsPermission(statErr) {
				return nil
			}
			return statErr
		}
		total += fi.Size()
		return nil
	})
	return total, err
}

func immediateChildSizes(dir string) ([]ChildLine, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var children []ChildLine
	for _, ent := range entries {
		childPath := filepath.Join(dir, ent.Name())
		size, sizeErr := deepSize(childPath)
		if sizeErr != nil {
			if os.IsNotExist(sizeErr) || os.IsPermission(sizeErr) {
				continue
			}
			return nil, sizeErr
		}
		children = append(children, ChildLine{
			Name:      ent.Name(),
			Bytes:     size,
			SizeHuman: FormatSize(size),
		})
	}
	sortChildren(children)
	return children, nil
}

func sortChildren(children []ChildLine) {
	for i := 0; i < len(children); i++ {
		for j := i + 1; j < len(children); j++ {
			if children[j].Name < children[i].Name {
				children[i], children[j] = children[j], children[i]
			}
		}
	}
}

func countNodeModulesDirs(root string) (int, error) {
	count := 0
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			if os.IsPermission(walkErr) {
				return filepath.SkipDir
			}
			return walkErr
		}
		if d.IsDir() && d.Name() == "node_modules" {
			count++
		}
		return nil
	})
	return count, err
}

func countTopLevelDirs(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, ent := range entries {
		if ent.IsDir() {
			n++
		}
	}
	return n
}

func countTopLevelFiles(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, ent := range entries {
		if !ent.IsDir() {
			n++
		}
	}
	return n
}

func countDirEntries(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	return len(entries)
}

func dirSizeIfExists(path string) int64 {
	size, err := deepSize(path)
	if err != nil {
		return 0
	}
	return size
}

func fileSizeAndLines(path string) (size int64, lines string, err error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, "", err
	}
	size = info.Size()
	_, isBinary, detectErr := detect.DetectFileType(path)
	if detectErr != nil || isBinary {
		return size, "(binary)", nil
	}
	n, err := countTextLines(path)
	if err != nil {
		return size, "", err
	}
	return size, fmt.Sprintf("%d", n), nil
}

func countTextLines(path string) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		lines++
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	if lines == 0 {
		return 0, nil
	}
	return lines, nil
}

func countRolloutFiles(sessionsDir string) int {
	count := 0
	_ = filepath.WalkDir(sessionsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		base := d.Name()
		if strings.HasPrefix(base, "rollout-") && strings.HasSuffix(base, ".jsonl") {
			count++
		}
		return nil
	})
	return count
}

func countSQLiteFamily(dir, prefix string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	seen := make(map[string]struct{})
	for _, ent := range entries {
		name := ent.Name()
		if !strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, ".sqlite") {
			continue
		}
		base := strings.TrimSuffix(name, ".sqlite")
		seen[base] = struct{}{}
	}
	return len(seen)
}