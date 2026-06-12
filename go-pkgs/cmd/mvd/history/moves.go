package history

func FindLastNonWorktreePath(locations []LocationEntry) string {
	for i := len(locations) - 1; i >= 0; i-- {
		loc := locations[i]
		if loc.Git == nil || loc.Git.Type != "worktree" {
			return loc.Path
		}
	}
	return locations[0].Path
}

func DeriveMoves(locs []LocationEntry) (string, []MoveEntry) {
	if len(locs) == 0 {
		return "", nil
	}
	root := locs[0].Path
	moves := make([]MoveEntry, 0, len(locs)-1)
	for i := 1; i < len(locs); i++ {
		loc := locs[i]
		move := MoveEntry{
			Current: loc.Path,
			Type:    "plain",
		}
		if loc.Git != nil && loc.Git.Type == "worktree" {
			move.Type = "worktree"
			move.Branch = loc.Git.Branch
			move.Prev = FindLastNonWorktreePath(locs[:i])
			moves = append(moves, move)
			continue
		}
		move.Prev = FindLastNonWorktreePath(locs[:i])
		moves = append(moves, move)
	}
	return root, moves
}

func LocationsFromMoves(root string, moves []MoveEntry) []LocationEntry {
	locs := []LocationEntry{{Path: root}}
	for _, move := range moves {
		loc := LocationEntry{Path: move.Current}
		if move.Type == "worktree" {
			loc.Git = &GitInfo{
				Type:     "worktree",
				MainRepo: move.Prev,
				Branch:   move.Branch,
			}
		}
		locs = append(locs, loc)
	}
	return locs
}
