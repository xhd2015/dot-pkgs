package singboxtun

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRuntimeLogPathsUnderPackageCacheDir(t *testing.T) {
	dir, err := packageCacheDir()
	if err != nil {
		t.Fatalf("packageCacheDir: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join(defaultCacheDirName)) {
		t.Fatalf("packageCacheDir = %q, want .../%s", dir, defaultCacheDirName)
	}
	path := singBoxLogPath()
	if !strings.HasPrefix(path, dir+string(os.PathSeparator)) {
		t.Fatalf("log path %q not under %q", path, dir)
	}
}

func TestActiveCacheDirNameOverride(t *testing.T) {
	prev := activeCacheDirName
	activeCacheDirName = "custom-cache"
	defer func() { activeCacheDirName = prev }()

	dir, err := packageCacheDir()
	if err != nil {
		t.Fatalf("packageCacheDir: %v", err)
	}
	if !strings.HasSuffix(dir, filepath.Join("custom-cache")) {
		t.Fatalf("packageCacheDir = %q, want .../custom-cache", dir)
	}
}
