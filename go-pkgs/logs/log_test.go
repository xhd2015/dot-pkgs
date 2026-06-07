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
		errCh <- WatchLine(ctx, path, WatchLineOptions{}, func(line string) error {
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
		errCh <- WatchLine(ctx, path, WatchLineOptions{}, func(line string) error {
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
		errCh <- WatchLine(ctx, path, WatchLineOptions{}, func(line string) error {
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

func TestWatchFileEventsNoDebounce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events := make(chan FileEvent, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchFileEvents(ctx, path, WatchFileEventsOptions{
			DisableDebounce: true,
		}, func(ev FileEvent) error {
			events <- ev
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)
	os.WriteFile(path, []byte("hello\n"), 0644)

	select {
	case ev := <-events:
		if ev.Op != FileCreated || ev.Path != path {
			t.Fatalf("got op=%v path=%s, want Created", ev.Op, ev.Path)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for create event")
	}

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("world\n")
	f.Sync()
	f.Close()

	select {
	case ev := <-events:
		if ev.Op != FileModified || ev.Path != path {
			t.Fatalf("got op=%v, want Modified", ev.Op)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for modify event")
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchFileEvents: %v", err)
	}
}

func TestWatchFileEventsDebounce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events := make(chan FileEvent, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchFileEvents(ctx, path, WatchFileEventsOptions{
			Debounce: 200 * time.Millisecond,
		}, func(ev FileEvent) error {
			events <- ev
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	os.WriteFile(path, []byte("x"), 0644)
	time.Sleep(10 * time.Millisecond)
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("y")
	f.Sync()
	f.Close()

	select {
	case ev := <-events:
		if ev.Op != FileCreated {
			t.Fatalf("got op=%v, want Created", ev.Op)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for coalesced event")
	}

	select {
	case <-events:
		t.Fatal("expected only 1 coalesced event")
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchFileEvents: %v", err)
	}
}

func TestWatchLineDebounce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lines := make(chan string, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchLine(ctx, path, WatchLineOptions{
			Debounce: 200 * time.Millisecond,
		}, func(line string) error {
			lines <- line
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	os.WriteFile(path, []byte("line1\n"), 0644)
	time.Sleep(10 * time.Millisecond)
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString("line2\n")
	f.Sync()
	f.Close()

	select {
	case line := <-lines:
		if line != "line1" {
			t.Fatalf("got %q, want line1", line)
		}
		select {
		case line2 := <-lines:
			if line2 != "line2" {
				t.Fatalf("got %q, want line2", line2)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("timeout waiting for line2")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for coalesced lines")
	}

	select {
	case <-lines:
		t.Fatal("expected only 2 lines from coalesced write")
	case <-time.After(100 * time.Millisecond):
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchLine: %v", err)
	}
}

func TestWatchLineDisableDebounce(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.log")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	lines := make(chan string, 10)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchLine(ctx, path, WatchLineOptions{
			DisableDebounce: true,
		}, func(line string) error {
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
		t.Fatal("timeout waiting for line")
	}

	cancel()
	if err := <-errCh; err != nil {
		t.Fatalf("WatchLine: %v", err)
	}
}
