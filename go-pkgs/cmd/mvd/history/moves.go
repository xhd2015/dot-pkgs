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

func locationType(loc LocationEntry) string {
	if loc.Git != nil && loc.Git.Type == "worktree" {
		return "worktree"
	}
	return "main"
}

func locationTypeAt(locs []LocationEntry, path string) string {
	for _, loc := range locs {
		if loc.Path == path {
			return locationType(loc)
		}
	}
	return "main"
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
			To:     loc.Path,
			ToType: locationType(loc),
		}
		if loc.Git != nil && loc.Git.Type == "worktree" {
			move.From = loc.Git.MainRepo
			move.Branch = loc.Git.Branch
		} else {
			move.From = FindLastNonWorktreePath(locs[:i])
		}
		move.FromType = locationTypeAt(locs[:i], move.From)
		moves = append(moves, move)
	}
	return root, moves
}

func LocationsFromMoves(root string, moves []MoveEntry) []LocationEntry {
	locs := []LocationEntry{{Path: root}}
	for _, move := range moves {
		loc := LocationEntry{Path: move.To}
		if move.ToType == "worktree" {
			loc.Git = &GitInfo{
				Type:     "worktree",
				MainRepo: move.From,
				Branch:   move.Branch,
			}
		}
		locs = append(locs, loc)
	}
	return locs
}