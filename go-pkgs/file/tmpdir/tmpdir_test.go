package tmpdir

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCommonTmpDir(t *testing.T) {
	fallback := "/system-temp"
	directory, err := os.Stat(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	notDirectory, err := os.Stat(file)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("non-Windows", func(t *testing.T) {
		t.Run("uses /tmp when it is a directory", func(t *testing.T) {
			got := getCommonTmpDir("linux", func(path string) (os.FileInfo, error) {
				if path != unixTmpDir {
					t.Fatalf("stat path=%q, want %q", path, unixTmpDir)
				}
				return directory, nil
			}, func() string { return fallback })
			if got != unixTmpDir {
				t.Fatalf("directory=%q, want %q", got, unixTmpDir)
			}
		})

		t.Run("falls back when /tmp is missing", func(t *testing.T) {
			got := getCommonTmpDir("darwin", func(string) (os.FileInfo, error) {
				return nil, errors.New("not found")
			}, func() string { return fallback })
			if got != fallback {
				t.Fatalf("directory=%q, want fallback %q", got, fallback)
			}
		})

		t.Run("falls back when /tmp is not a directory", func(t *testing.T) {
			got := getCommonTmpDir("freebsd", func(string) (os.FileInfo, error) {
				return notDirectory, nil
			}, func() string { return fallback })
			if got != fallback {
				t.Fatalf("directory=%q, want fallback %q", got, fallback)
			}
		})
	})

	t.Run("Windows uses the operating system temporary directory", func(t *testing.T) {
		statCalled := false
		got := getCommonTmpDir("windows", func(string) (os.FileInfo, error) {
			statCalled = true
			return nil, nil
		}, func() string { return fallback })
		if statCalled {
			t.Fatal("Windows must not probe /tmp")
		}
		if got != fallback {
			t.Fatalf("directory=%q, want fallback %q", got, fallback)
		}
	})
}
