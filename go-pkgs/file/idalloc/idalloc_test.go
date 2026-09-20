package idalloc

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
)

func TestInferLockFile(t *testing.T) {
	cases := []struct{ state, want string }{
		{"/data/id.json", "/data/id.lock"},
		{"/data/ids", "/data/ids.lock"},
		{"/data/ids.v2.json", "/data/ids.v2.lock"},
		{"/data/.hidden.json", "/data/.hidden.lock"},
	}
	for _, c := range cases {
		if got := inferLockFile(c.state); got != c.want {
			t.Errorf("inferLockFile(%q) = %q, want %q", c.state, got, c.want)
		}
	}
}

func TestNewCreatesLockFileWithInferredName(t *testing.T) {
	dir := t.TempDir()
	a := New(filepath.Join(dir, "ids"))
	if _, err := a.Next(0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "ids.lock")); err != nil {
		t.Fatalf("inferred lock file missing: %v", err)
	}
}

func TestNextSequential(t *testing.T) {
	a := New(filepath.Join(t.TempDir(), "id.json"))
	for want := int64(1); want <= 3; want++ {
		got, err := a.Next(0)
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("Next = %d, want %d", got, want)
		}
	}
}

func TestNextFloorAndPersistence(t *testing.T) {
	a := New(filepath.Join(t.TempDir(), "id.json"))
	// Floor raises a missing state file past existing data.
	if got, err := a.Next(42); err != nil || got != 43 {
		t.Fatalf("Next(42) = %d, %v; want 43", got, err)
	}
	// A persisted last value beats a lower floor.
	if got, err := a.Next(5); err != nil || got != 44 {
		t.Fatalf("Next(5) = %d, %v; want 44", got, err)
	}
	// A fresh allocator instance continues the sequence (restart).
	if got, err := New(filepath.Join(filepath.Dir(a.stateFile), "id.json")).Next(0); err != nil || got != 45 {
		t.Fatalf("fresh allocator Next = %d, %v; want 45", got, err)
	}
}

func TestLast(t *testing.T) {
	a := New(filepath.Join(t.TempDir(), "id.json"))
	if got, err := a.Last(); err != nil || got != 0 {
		t.Fatalf("Last on missing state = %d, %v; want 0", got, err)
	}
	if _, err := a.Next(0); err != nil {
		t.Fatal(err)
	}
	got, err := a.Last()
	if err != nil || got != 1 {
		t.Fatalf("Last = %d, %v; want 1", got, err)
	}
	// Last must not allocate.
	again, err := a.Last()
	if err != nil || again != 1 {
		t.Fatalf("second Last = %d, %v; want 1", again, err)
	}
}

func TestNextConcurrent(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "id.json")
	const goroutines, each = 8, 25
	var mu sync.Mutex
	seen := map[int64]bool{}
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// Shared object and separate objects on the same state file
			// must both serialize correctly.
			a := New(stateFile)
			for i := 0; i < each; i++ {
				n, err := a.Next(0)
				if err != nil {
					t.Error(err)
					return
				}
				mu.Lock()
				if seen[n] {
					t.Errorf("duplicate id %d", n)
				}
				seen[n] = true
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if len(seen) != goroutines*each {
		t.Fatalf("allocated %d unique ids, want %d", len(seen), goroutines*each)
	}
}

func TestCorruptState(t *testing.T) {
	dir := t.TempDir()
	stateFile := filepath.Join(dir, "id.json")
	if err := os.WriteFile(stateFile, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	a := New(stateFile)
	if _, err := a.Next(0); err == nil || !strings.Contains(err.Error(), "reseed") {
		t.Fatalf("Next error = %v, want reseed hint", err)
	}
	if _, err := a.Last(); err == nil || !strings.Contains(err.Error(), "reseed") {
		t.Fatalf("Last error = %v, want reseed hint", err)
	}
}

func TestFilePermissions(t *testing.T) {
	root := filepath.Join(t.TempDir(), "nested", "data")
	a := New(filepath.Join(root, "id.json"))
	if _, err := a.Next(0); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"id.json", "id.lock"} {
		info, err := os.Stat(filepath.Join(root, name))
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("%s mode %o, want 600", name, info.Mode().Perm())
		}
	}
	info, err := os.Stat(root)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o700 {
		t.Errorf("dir mode %o, want 700", info.Mode().Perm())
	}
}

func TestLockFileIsNeverRenamed(t *testing.T) {
	dir := t.TempDir()
	a := New(filepath.Join(dir, "id.json"))
	if _, err := a.Next(0); err != nil {
		t.Fatal(err)
	}
	lockIno := inode(t, filepath.Join(dir, "id.lock"))
	stateIno := inode(t, filepath.Join(dir, "id.json"))
	for i := 0; i < 3; i++ {
		if _, err := a.Next(0); err != nil {
			t.Fatal(err)
		}
	}
	if got := inode(t, filepath.Join(dir, "id.lock")); got != lockIno {
		t.Errorf("lock file inode changed: %d -> %d", lockIno, got)
	}
	if got := inode(t, filepath.Join(dir, "id.json")); got == stateIno {
		t.Errorf("state file was replaced in place, want atomic rename")
	}
}

func inode(t *testing.T, path string) uint64 {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Skip("no raw stat on this platform")
	}
	return stat.Ino
}

func TestMaxNumeric(t *testing.T) {
	got := MaxNumeric([]string{"aabbcc", "12", "007", "999", "", "not-a-number"})
	if got != 999 {
		t.Fatalf("MaxNumeric = %d, want 999", got)
	}
	if got := MaxNumeric(nil); got != 0 {
		t.Fatalf("MaxNumeric(nil) = %d, want 0", got)
	}
}

func TestMissingParentDirIsCreated(t *testing.T) {
	root := filepath.Join(t.TempDir(), "does", "not", "exist")
	a := New(filepath.Join(root, "id.json"))
	got, err := a.Next(0)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("Next = %d, want 1", got)
	}
	if _, err := os.Stat(filepath.Join(root, "id.json")); errors.Is(err, os.ErrNotExist) {
		t.Fatal("state file not written")
	}
}
