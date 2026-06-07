package logs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestWatchLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lines := make(chan string, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchLine(ctx, path, func(line string) error {
			lines <- line
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	os.WriteFile(path, []byte("hello\n"), 0644)
	select {
	case line := <-lines:
		if line != "hello" {
			t.Fatalf("got %q, want hello", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for first line")
	}

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("world\n")
	f.WriteString("foo\nbar\n")
	f.Sync()
	f.Close()

	for _, want := range []string{"world", "foo", "bar"} {
		select {
		case line := <-lines:
			if line != want {
				t.Fatalf("got %q, want %q", line, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for %q", want)
		}
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchLine: %v", err)
	}
}

func TestWatchLineHandlesIncompleteLine(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lines := make(chan string, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchLine(ctx, path, func(line string) error {
			lines <- line
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	os.WriteFile(path, []byte("hel"), 0644)

	time.Sleep(200 * time.Millisecond)
	select {
	case <-lines:
		t.Fatal("should not receive incomplete line")
	case <-time.After(200 * time.Millisecond):
	}

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("lo\n")
	f.Sync()
	f.Close()

	select {
	case line := <-lines:
		if line != "hello" {
			t.Fatalf("got %q, want hello", line)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for completed line")
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchLine: %v", err)
	}
}

func TestWatchLineCarriageReturn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lines := make(chan string, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchLine(ctx, path, func(line string) error {
			lines <- line
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	os.WriteFile(path, []byte("hello\r\nworld\r\n"), 0644)

	for _, want := range []string{"hello", "world"} {
		select {
		case line := <-lines:
			if line != want {
				t.Fatalf("got %q, want %q", line, want)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("timeout waiting for %q", want)
		}
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchLine: %v", err)
	}
}
