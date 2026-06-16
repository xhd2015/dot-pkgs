package history

import (
	"testing"
)

func TestDeriveMovesWorktreeFromMain(t *testing.T) {
	root := "/proj/repo"
	wt := "/proj/wt"
	locs := []LocationEntry{
		{Path: root},
		{Path: wt, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature"}},
	}

	_, moves := DeriveMoves(locs)
	if len(moves) != 1 {
		t.Fatalf("expected 1 move, got %d", len(moves))
	}
	m := moves[0]
	if m.From != root || m.FromType != "main" || m.To != wt || m.ToType != "worktree" || m.Branch != "feature" {
		t.Fatalf("unexpected move: %#v", m)
	}
}

func TestDeriveMovesPlainAfterWorktree(t *testing.T) {
	root := "/proj/repo"
	wt := "/proj/wt"
	dst := "/proj/dst"
	locs := []LocationEntry{
		{Path: root},
		{Path: wt, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "feature"}},
		{Path: dst},
	}

	_, moves := DeriveMoves(locs)
	if len(moves) != 2 {
		t.Fatalf("expected 2 moves, got %d", len(moves))
	}
	plain := moves[1]
	if plain.From != root || plain.FromType != "main" || plain.To != dst || plain.ToType != "main" {
		t.Fatalf("unexpected plain move: %#v", plain)
	}
}

func TestDeriveMovesWorktreeFromExternalMain(t *testing.T) {
	root := "/proj/repo"
	wt1 := "/proj/wt1"
	dst := "/proj/dst"
	wt2 := "/proj/wt2"
	locs := []LocationEntry{
		{Path: root},
		{Path: wt1, Git: &GitInfo{Type: "worktree", MainRepo: root, Branch: "a"}},
		{Path: dst},
		{Path: wt2, Git: &GitInfo{Type: "worktree", MainRepo: dst, Branch: "b"}},
	}

	_, moves := DeriveMoves(locs)
	if len(moves) != 3 {
		t.Fatalf("expected 3 moves, got %d", len(moves))
	}
	ext := moves[2]
	if ext.From != dst || ext.FromType != "main" || ext.To != wt2 || ext.ToType != "worktree" {
		t.Fatalf("unexpected external-main worktree move: %#v", ext)
	}
}

func TestLocationsFromMovesRoundTrip(t *testing.T) {
	locs := []LocationEntry{
		{Path: "/proj/repo"},
		{Path: "/proj/wt", Git: &GitInfo{Type: "worktree", MainRepo: "/proj/repo", Branch: "feature"}},
		{Path: "/proj/dst"},
		{Path: "/proj/wt2", Git: &GitInfo{Type: "worktree", MainRepo: "/proj/dst", Branch: "other"}},
	}

	root, moves := DeriveMoves(locs)
	got := LocationsFromMoves(root, moves)
	if len(got) != len(locs) {
		t.Fatalf("length mismatch: got %d want %d", len(got), len(locs))
	}
	for i := range locs {
		if got[i].Path != locs[i].Path {
			t.Fatalf("path[%d]: got %s want %s", i, got[i].Path, locs[i].Path)
		}
		if (got[i].Git == nil) != (locs[i].Git == nil) {
			t.Fatalf("git[%d]: got %#v want %#v", i, got[i].Git, locs[i].Git)
		}
		if got[i].Git != nil {
			if got[i].Git.MainRepo != locs[i].Git.MainRepo || got[i].Git.Branch != locs[i].Git.Branch {
				t.Fatalf("git[%d]: got %#v want %#v", i, got[i].Git, locs[i].Git)
			}
		}
	}
}