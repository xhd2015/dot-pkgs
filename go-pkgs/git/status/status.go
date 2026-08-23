package status

import (
	"fmt"
	"strings"

	gitopsStatus "github.com/xhd2015/gitops/git/status"
)

type Counts struct {
	Modified, Added, Deleted, Untracked, Renamed, Copied, Unmerged int
}

type FormatStyle int

const (
	FormatBackup FormatStyle = iota
	StyleWrk
)

// WrkCounts is the wrk five-bucket view (distinct from backup Counts labels).
// Each porcelain line increments exactly one bucket (path-once).
// Any index (staged) change counts as Staged only — not also Changed/Renamed/Deleted.
type WrkCounts struct {
	Staged, Changed, Renamed, Deleted, Untracked int
}

func (c WrkCounts) dirty() bool {
	return c.Staged+c.Changed+c.Renamed+c.Deleted+c.Untracked > 0
}

func ParsePorcelain(porcelain string) Counts {
	c := gitopsStatus.ParsePorcelain(porcelain)
	return Counts{
		Modified:  c.Modified,
		Added:     c.Added,
		Deleted:   c.Deleted,
		Untracked: c.Untracked,
		Renamed:   c.Renamed,
		Copied:    c.Copied,
		Unmerged:  c.Unmerged,
	}
}

// ParsePorcelainWrk applies wrk taxonomy (path-once, first match wins):
//
//	?? → Untracked
//	index column non-blank (not '?') → Staged (covers A/M/D/R/… including AM;
//	  staged paths are not also Changed/Renamed/Deleted)
//	else worktree R → Renamed
//	else worktree D → Deleted
//	else → Changed
func ParsePorcelainWrk(porcelain string) WrkCounts {
	var counts WrkCounts
	for _, line := range strings.Split(porcelain, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			counts.Untracked++
			continue
		}
		if len(line) < 2 {
			counts.Changed++
			continue
		}
		x, y := line[0], line[1]
		if x != ' ' && x != '?' {
			counts.Staged++
			continue
		}
		switch {
		case y == 'R':
			counts.Renamed++
		case y == 'D':
			counts.Deleted++
		default:
			counts.Changed++
		}
	}
	return counts
}

// FormatWrk renders wrk --status Status: value (no ANSI).
// Deprecated: wrk owns this wording in wrkcli (formatWrkStatus). Kept as a shim for
// checkout.Enrich (StyleWrk) and other non-wrk callers; wrk should not call FormatWrk.
func FormatWrk(counts WrkCounts) string {
	if !counts.dirty() {
		return "clean"
	}
	return fmt.Sprintf("dirty (%d staged, %d changed, %d renamed, %d deleted, %d untracked)",
		counts.Staged, counts.Changed, counts.Renamed, counts.Deleted, counts.Untracked)
}

func (c Counts) dirty() bool {
	return c.Modified+c.Added+c.Deleted+c.Untracked+c.Renamed+c.Copied+c.Unmerged > 0
}

func Format(counts Counts, style FormatStyle) string {
	switch style {
	case FormatBackup:
		if !counts.dirty() {
			return "clean"
		}
		var parts []string
		if counts.Modified > 0 {
			parts = append(parts, fmt.Sprintf("%d modified", counts.Modified))
		}
		if counts.Added > 0 {
			parts = append(parts, fmt.Sprintf("%d added", counts.Added))
		}
		if counts.Deleted > 0 {
			parts = append(parts, fmt.Sprintf("%d deleted", counts.Deleted))
		}
		if counts.Untracked > 0 {
			parts = append(parts, fmt.Sprintf("%d untracked", counts.Untracked))
		}
		if counts.Renamed > 0 {
			parts = append(parts, fmt.Sprintf("%d renamed", counts.Renamed))
		}
		if counts.Copied > 0 {
			parts = append(parts, fmt.Sprintf("%d copied", counts.Copied))
		}
		if counts.Unmerged > 0 {
			parts = append(parts, fmt.Sprintf("%d unmerged", counts.Unmerged))
		}
		return "dirty (" + strings.Join(parts, ", ") + ")"
	case StyleWrk:
		return FormatWrk(wrkCountsFromBackup(counts))
	default:
		return Format(counts, FormatBackup)
	}
}

func wrkCountsFromBackup(counts Counts) WrkCounts {
	return WrkCounts{
		Staged:    counts.Added,
		Changed:   counts.Modified,
		Renamed:   counts.Renamed,
		Deleted:   counts.Deleted,
		Untracked: counts.Untracked,
	}
}
