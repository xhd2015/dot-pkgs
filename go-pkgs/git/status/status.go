package status

import (
	"fmt"
	"strings"
)

type Counts struct {
	Modified, Added, Deleted, Untracked, Renamed, Copied, Unmerged int
}

type FormatStyle int

const (
	FormatBackup FormatStyle = iota
	StyleWrk
)

// WrkCounts is the wrk four-bucket view (distinct from backup Counts labels).
type WrkCounts struct {
	Added, Changed, Renamed, Deleted int
}

func (c WrkCounts) dirty() bool {
	return c.Added+c.Changed+c.Renamed+c.Deleted > 0
}

func ParsePorcelain(porcelain string) Counts {
	var counts Counts
	for _, line := range strings.Split(porcelain, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if len(line) < 2 {
			continue
		}
		x, y := line[0], line[1]
		if x == '?' && y == '?' {
			counts.Untracked++
			continue
		}
		if x == 'R' {
			counts.Renamed++
			continue
		}
		if x == 'C' {
			counts.Copied++
			continue
		}
		if x == 'A' {
			counts.Added++
			continue
		}
		if x == 'U' || y == 'U' {
			counts.Unmerged++
		}
		if x == 'D' || y == 'D' {
			counts.Deleted++
			continue
		}
		if x == 'M' || y == 'M' {
			counts.Modified++
		}
	}
	return counts
}

// ParsePorcelainWrk applies wrk taxonomy (?? → added; M/default → changed).
func ParsePorcelainWrk(porcelain string) WrkCounts {
	var counts WrkCounts
	for _, line := range strings.Split(porcelain, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "??") {
			counts.Added++
			continue
		}
		if len(line) < 2 {
			counts.Changed++
			continue
		}
		x, y := line[0], line[1]
		switch {
		case x == 'R' || y == 'R':
			counts.Renamed++
		case x == 'A' || y == 'A':
			counts.Added++
		case x == 'D' || y == 'D':
			counts.Deleted++
		default:
			counts.Changed++
		}
	}
	return counts
}

// FormatWrk renders wrk --status Status: value (no ANSI).
func FormatWrk(counts WrkCounts) string {
	if !counts.dirty() {
		return "clean"
	}
	return fmt.Sprintf("dirty (%d added, %d changed, %d renamed, %d deleted)",
		counts.Added, counts.Changed, counts.Renamed, counts.Deleted)
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
		Added:    counts.Added + counts.Untracked,
		Changed:  counts.Modified,
		Renamed:  counts.Renamed,
		Deleted:  counts.Deleted,
	}
}