package server

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDirectoryControlStopsOnlyMatchingDirectory(t *testing.T) {
	root := t.TempDir()
	other := t.TempDir()
	first, err := openDirectoryControl(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := first.start(cancel); err != nil {
		first.unlock()
		t.Fatal(err)
	}
	first.unlock()
	closed := false
	defer func() {
		if !closed {
			first.close()
		}
	}()
	otherControl, err := openDirectoryControl(other)
	if err != nil {
		t.Fatal(err)
	}
	found, err := otherControl.stopExisting(Config{Root: other, Stdout: io.Discard})
	otherControl.unlock()
	if err != nil || found {
		t.Fatalf("other directory: found=%v err=%v", found, err)
	}
	select {
	case <-ctx.Done():
		t.Fatal("stopped unrelated directory")
	default:
	}
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		second, err := openDirectoryControl(alias)
		if err != nil {
			done <- err
			return
		}
		defer second.unlock()
		found, err := second.stopExisting(Config{Root: alias, Stdout: io.Discard})
		if err == nil && !found {
			err = os.ErrNotExist
		}
		done <- err
	}()
	select {
	case <-ctx.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("stop not delivered")
	}
	// Acknowledge only after the supervisor has finished cleaning children.
	first.close()
	closed = true
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("stop not acknowledged")
	}
}
