package logs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWatchCreateMatch_nestedCreate(t *testing.T) {
	root := t.TempDir()
	sessions := filepath.Join(root, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	got := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- WatchCreateMatch(ctx, sessions, WatchCreateMatchOptions{MaxDepth: 4},
			func(path string) bool {
				base := filepath.Base(path)
				return strings.HasPrefix(base, "rollout-") && strings.Contains(base, "abc123")
			},
			func(path string) error {
				select {
				case got <- path:
				default:
				}
				cancel()
				return nil
			},
		)
	}()

	time.Sleep(80 * time.Millisecond)
	// Create one level at a time so each directory Create can be hydrated.
	for _, seg := range []string{"2026", "09", "03"} {
		sessions = filepath.Join(sessions, seg)
		if err := os.Mkdir(sessions, 0o755); err != nil {
			t.Fatal(err)
		}
		time.Sleep(30 * time.Millisecond)
	}
	day := sessions
	target := filepath.Join(day, "rollout-2026-09-03T12-00-00-abc123.jsonl")
	if err := os.WriteFile(target, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	select {
	case p := <-got:
		if p != target {
			// macOS may resolve slightly; require suffix match
			if !strings.HasSuffix(p, filepath.Base(target)) {
				t.Fatalf("got %q want %q", p, target)
			}
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for create match")
	}
	<-errCh
}

func TestWatchCreateMatch_existingFile(t *testing.T) {
	root := t.TempDir()
	sessions := filepath.Join(root, "sessions", "2026", "09", "03")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(sessions, "rollout-x-abc123.jsonl")
	if err := os.WriteFile(target, []byte("{}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	got := make(chan string, 1)
	go func() {
		_ = WatchCreateMatch(ctx, filepath.Join(root, "sessions"), WatchCreateMatchOptions{},
			func(path string) bool { return strings.Contains(filepath.Base(path), "abc123") },
			func(path string) error {
				got <- path
				cancel()
				return nil
			},
		)
	}()
	select {
	case <-got:
	case <-time.After(2 * time.Second):
		t.Fatal("expected existing match")
	}
}
