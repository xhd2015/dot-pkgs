package ocr

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHelperNameStable(t *testing.T) {
	src := []byte("same source")
	a := helperName(src)
	b := helperName(src)
	if a != b {
		t.Fatalf("helperName unstable: %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, helperPrefix) {
		t.Fatalf("helperName %q missing prefix %q", a, helperPrefix)
	}
	if helperName([]byte("other")) == a {
		t.Fatal("different source should change helper name")
	}
}

func TestResolveCacheDirExplicit(t *testing.T) {
	got, err := resolveCacheDir(Options{CacheDir: "/tmp/custom-ocr-cache"})
	if err != nil {
		t.Fatal(err)
	}
	if got != "/tmp/custom-ocr-cache" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveCacheDirDefault(t *testing.T) {
	got, err := resolveCacheDir(Options{
		UserCacheDir: func() (string, error) { return "/tmp/user-cache", nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join("/tmp/user-cache", cacheSubdir)
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestWithDirLockSerializes(t *testing.T) {
	dir := t.TempDir()
	lock := filepath.Join(dir, "lock")
	started := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)

	go func() {
		done <- withDirLock(lock, func() error {
			close(started)
			<-release
			return nil
		})
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("first lock holder did not start")
	}

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- withDirLock(lock, func() error {
			return nil
		})
	}()

	select {
	case <-secondDone:
		t.Fatal("second lock acquired while first held")
	case <-time.After(150 * time.Millisecond):
	}

	close(release)
	if err := <-done; err != nil {
		t.Fatalf("first lock: %v", err)
	}
	select {
	case err := <-secondDone:
		if err != nil {
			t.Fatalf("second lock: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("second lock did not complete")
	}
}

func TestWithDirLockCleansUp(t *testing.T) {
	dir := t.TempDir()
	lock := filepath.Join(dir, "lock")
	if err := withDirLock(lock, func() error { return nil }); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(lock); !os.IsNotExist(err) {
		t.Fatalf("lock dir should be removed, stat err=%v", err)
	}
}
