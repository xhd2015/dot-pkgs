package analyse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/xhd2015/dot-pkgs/go-pkgs/git/scan_repo"
)

// ResolveHome returns the absolute path to HOME after validation.
func ResolveHome(home string) (string, error) {
	return resolveHome(home)
}

func resolveHome(home string) (string, error) {
	if home == "" {
		home = os.Getenv("HOME")
	}
	if home == "" {
		return "", fmt.Errorf("HOME is not set")
	}
	abs, err := filepath.Abs(home)
	if err != nil {
		return "", fmt.Errorf("resolve home: %w", err)
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", fmt.Errorf("stat home: %w", err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("HOME is not a directory: %s", abs)
	}
	return abs, nil
}

// Scan walks every immediate child of HOME and returns per-entry results and summary.
func Scan(ctx context.Context, opts Options) ([]EntryResult, ScanSummary, error) {
	home, err := resolveHome(opts.Home)
	if err != nil {
		return nil, ScanSummary{}, err
	}

	entries, err := os.ReadDir(home)
	if err != nil {
		return nil, ScanSummary{}, fmt.Errorf("read home: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	summary := ScanSummary{Home: home}
	var results []EntryResult

	for _, ent := range entries {
		entryPath := filepath.Join(home, ent.Name())
		var result EntryResult
		if ent.IsDir() {
			result, err = scanDirEntry(ctx, ent.Name(), entryPath)
		} else {
			result, err = scanFileEntry(ent.Name(), entryPath)
		}
		if err != nil {
			return results, summary, fmt.Errorf("scan %s: %w", ent.Name(), err)
		}

		summary.EntryCount++
		if ent.IsDir() {
			summary.DirCount++
		} else {
			summary.FileCount++
		}
		summary.TotalBytes += result.Bytes
		summary.GitRepos += result.Aggregates.GitRepos
		summary.LinkedWorktrees += result.Aggregates.LinkedWorktrees
		summary.NodeModulesDirs += result.Aggregates.NodeModulesDirs

		switch ent.Name() {
		case ".codex":
			summary.HasCodex = true
			tc := topicCountsFromEntry(ent.Name(), entryPath)
			summary.CodexSessions = tc.CodexSessions
			summary.CodexSkills = tc.CodexSkills
		case ".grok":
			summary.HasGrok = true
			tc := topicCountsFromEntry(ent.Name(), entryPath)
			summary.GrokSessions = tc.GrokSessions
			summary.GrokProjects = tc.GrokProjects
			summary.GrokSkills = tc.GrokSkills
		case ".cursor":
			summary.HasCursor = true
			tc := topicCountsFromEntry(ent.Name(), entryPath)
			summary.CursorProjects = tc.CursorProjects
			summary.CursorChats = tc.CursorChats
		case ".knowledge-hub":
			summary.HasKnowledgeHub = true
			tc := topicCountsFromEntry(ent.Name(), entryPath)
			summary.KHKnowledges = tc.KHKnowledges
		case ".knowledge-index":
			summary.HasKnowledgeIndex = true
			tc := topicCountsFromEntry(ent.Name(), entryPath)
			summary.KIAgents = tc.KIAgents
		case ".openclaw":
			summary.HasOpenclaw = true
			tc := topicCountsFromEntry(ent.Name(), entryPath)
			summary.OpenclawAgents = tc.OpenclawAgents
		}

		if opts.OnEntry != nil {
			if err := opts.OnEntry(result); err != nil {
				return results, summary, err
			}
		}
		results = append(results, result)
	}

	summary.TotalHuman = FormatSize(summary.TotalBytes)
	summary.Largest = topEntries(results, 5)
	return results, summary, nil
}

func scanFileEntry(name, path string) (EntryResult, error) {
	size, lines, err := fileSizeAndLines(path)
	if err != nil {
		return EntryResult{}, err
	}
	return EntryResult{
		Name:      name,
		Kind:      EntryKindFile,
		Bytes:     size,
		SizeHuman: FormatSize(size),
		Lines:     lines,
	}, nil
}

// ScanDirEntry scans one directory and returns analyse-files style entry data.
func ScanDirEntry(ctx context.Context, name, path string) (EntryResult, error) {
	return scanDirEntry(ctx, name, path)
}

func scanDirEntry(ctx context.Context, name, path string) (EntryResult, error) {
	total, err := deepSize(path)
	if err != nil {
		return EntryResult{}, err
	}

	children, err := immediateChildSizes(path)
	if err != nil {
		return EntryResult{}, err
	}

	semantic := semanticLinesForEntry(name, path)

	gitRepos, linked, err := gitAggregates(ctx, path)
	if err != nil {
		return EntryResult{}, err
	}

	nmDirs, err := countNodeModulesDirs(path)
	if err != nil {
		return EntryResult{}, err
	}

	return EntryResult{
		Name:      name,
		Kind:      EntryKindDir,
		Bytes:     total,
		SizeHuman: FormatSize(total),
		Children:  children,
		Semantic:  semantic,
		Aggregates: Aggregates{
			GitRepos:        gitRepos,
			LinkedWorktrees: linked,
			NodeModulesDirs: nmDirs,
		},
	}, nil
}

func gitAggregates(ctx context.Context, root string) (repos int, linked int, err error) {
	_, err = scan_repo.Scan(ctx, scan_repo.Options{
		Roots:         []string{root},
		ListWorktrees: true,
		OnRepo: func(repo scan_repo.Repo) error {
			repos++
			for _, wt := range repo.Worktrees {
				if !wt.IsMain {
					linked++
				}
			}
			return nil
		},
	})
	return repos, linked, err
}

func topEntries(results []EntryResult, n int) []LargestEntry {
	type pair struct {
		name  string
		bytes int64
	}
	var ranked []pair
	for _, r := range results {
		ranked = append(ranked, pair{name: r.Name, bytes: r.Bytes})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].bytes != ranked[j].bytes {
			return ranked[i].bytes > ranked[j].bytes
		}
		return ranked[i].name < ranked[j].name
	})
	if len(ranked) > n {
		ranked = ranked[:n]
	}
	out := make([]LargestEntry, len(ranked))
	for i, r := range ranked {
		out[i] = LargestEntry{
			Name:      r.name,
			Bytes:     r.bytes,
			SizeHuman: FormatSize(r.bytes),
		}
	}
	return out
}