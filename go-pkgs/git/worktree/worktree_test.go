package worktree

import "testing"

func TestParsePorcelain(t *testing.T) {
	input := `worktree /path/to/repo
HEAD 1234567890abcdef
branch refs/heads/main

worktree /path/to/repo/feature
HEAD 1234567890abcdef
branch refs/heads/feature
`

	entries := parsePorcelain(input)
	if len(entries) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(entries))
	}
	if entries[0].Path != "/path/to/repo" {
		t.Fatalf("expected first worktree to be /path/to/repo, got %s", entries[0].Path)
	}
	if entries[0].Branch != "main" {
		t.Fatalf("expected first branch main, got %s", entries[0].Branch)
	}
	if entries[1].Path != "/path/to/repo/feature" {
		t.Fatalf("expected second worktree to be /path/to/repo/feature, got %s", entries[1].Path)
	}
	if entries[1].Branch != "feature" {
		t.Fatalf("expected second branch feature, got %s", entries[1].Branch)
	}
}

func TestParsePorcelainEmpty(t *testing.T) {
	entries := parsePorcelain("")
	if len(entries) != 0 {
		t.Fatalf("expected 0 worktrees, got %d", len(entries))
	}
}

func TestParsePorcelainWithBare(t *testing.T) {
	input := `worktree /path/to/bare
bare

worktree /path/to/repo
HEAD abc
branch refs/heads/main
`
	entries := parsePorcelain(input)
	if len(entries) != 2 {
		t.Fatalf("expected 2 worktrees, got %d", len(entries))
	}
}