package withgo

import (
	"path/filepath"
	"testing"
)

func TestTargetGorootDoesNotRequireInstalledSDK(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), "installed")
	got := TargetGoroot("go1.19", ResolveOptions{InstallDir: installDir})
	want := filepath.Join(installDir, "go1.19.13")
	if got != want {
		t.Fatalf("TargetGoroot() = %q, want %q", got, want)
	}
}
