package submodule

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestDetectSubModules(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(dir string) error
		files    []string
		expected []string
	}{
		{
			name:     "empty files",
			setup:    nil,
			files:    nil,
			expected: nil,
		},
		{
			name:  "no submodules",
			setup: nil,
			files: []string{"src/main.go", "README.md"},
		},
		{
			name: "direct submodule directory with .git dir",
			setup: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755)
			},
			files:    []string{"vendor/libfoo"},
			expected: []string{"vendor/libfoo"},
		},
		{
			name: "direct submodule directory with .git file (worktree)",
			setup: func(dir string) error {
				smDir := filepath.Join(dir, "vendor", "libfoo")
				if err := os.MkdirAll(smDir, 0755); err != nil {
					return err
				}
				return os.WriteFile(filepath.Join(smDir, ".git"), []byte("gitdir: ../../.git/modules/vendor/libfoo"), 0644)
			},
			files:    []string{"vendor/libfoo"},
			expected: []string{"vendor/libfoo"},
		},
		{
			name: "file inside submodule",
			setup: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755)
			},
			files:    []string{"vendor/libfoo/src/main.c", "vendor/libfoo/README.md"},
			expected: []string{"vendor/libfoo"},
		},
		{
			name: "multiple files from same submodule",
			setup: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755)
			},
			files:    []string{"vendor/libfoo/src/main.c", "vendor/libfoo/src/util.c"},
			expected: []string{"vendor/libfoo"},
		},
		{
			name: "multiple submodules",
			setup: func(dir string) error {
				if err := os.MkdirAll(filepath.Join(dir, "vendor", "libfoo", ".git"), 0755); err != nil {
					return err
				}
				return os.MkdirAll(filepath.Join(dir, "third_party", "libbar", ".git"), 0755)
			},
			files:    []string{"vendor/libfoo/src/main.c", "third_party/libbar/src/main.c", "src/app.go"},
			expected: []string{"third_party/libbar", "vendor/libfoo"},
		},
		{
			name: "regular files only (no .git in parents)",
			setup: func(dir string) error {
				return os.MkdirAll(filepath.Join(dir, "src", "internal"), 0755)
			},
			files: []string{"src/main.go", "src/internal/util.go"},
		},
		{
			name: "non-existent files",
			files: []string{"nonexistent/file.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var workDir string
			if tt.setup != nil {
				tmpDir, err := os.MkdirTemp("", "submodule-test-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(tmpDir)
				workDir = tmpDir

				oldDir, _ := os.Getwd()
				os.Chdir(workDir)
				defer os.Chdir(oldDir)

				if err := tt.setup(workDir); err != nil {
					t.Fatal(err)
				}
			} else {
				var err error
				workDir, err = os.MkdirTemp("", "submodule-test-*")
				if err != nil {
					t.Fatal(err)
				}
				defer os.RemoveAll(workDir)
			}

			result := DetectSubModules(tt.files)
			sort.Strings(result)
			sort.Strings(tt.expected)

			if len(result) != len(tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("expected %v, got %v", tt.expected, result)
					return
				}
			}
		})
	}
}
