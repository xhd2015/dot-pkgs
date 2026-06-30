package remotefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsRemoteBackedPath(t *testing.T) {
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	if err := os.MkdirAll(home, 0755); err != nil {
		t.Fatal(err)
	}
	localRepo := filepath.Join(home, "Projects", "app")
	if err := os.MkdirAll(localRepo, 0755); err != nil {
		t.Fatal(err)
	}
	cloudRoot := filepath.Join(home, "Library", "CloudStorage", "GoogleDrive-user@example.com")
	if err := os.MkdirAll(cloudRoot, 0755); err != nil {
		t.Fatal(err)
	}

	t.Run("local path", func(t *testing.T) {
		remote, err := IsRemoteBackedPath(localRepo)
		if err != nil {
			t.Fatal(err)
		}
		if remote {
			t.Fatal("local repo should not be remote-backed")
		}
	})

	t.Run("cloud storage path", func(t *testing.T) {
		remote, err := IsRemoteBackedPath(cloudRoot)
		if err != nil {
			t.Fatal(err)
		}
		if !remote {
			t.Fatal("cloud storage provider should be remote-backed")
		}
	})

	t.Run("missing path", func(t *testing.T) {
		_, err := IsRemoteBackedPath(filepath.Join(home, "missing"))
		if err == nil {
			t.Fatal("expected stat error for missing path")
		}
	})

	t.Run("empty path", func(t *testing.T) {
		_, err := IsRemoteBackedPath("")
		if err == nil {
			t.Fatal("expected error for empty path")
		}
	})
}