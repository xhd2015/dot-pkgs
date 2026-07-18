package scan_repo

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// UniverseHome is the P2 library universe for single-root CacheRoot scans.
const UniverseHome = "home"

// seedHomeRepoIndex writes universe "home" repos.json for repos discovered under absRoot.
// Entries outside absRoot from a prior index are preserved (multi-root merge).
// No-op when cacheRoot is empty.
func seedHomeRepoIndex(cacheRoot, absRoot string, repos []Repo) error {
	if cacheRoot == "" {
		return nil
	}
	now := time.Now().UTC().Format(time.RFC3339)
	entries := make([]RepoIndexEntry, 0, len(repos))

	// Keep entries from other roots when merging into the shared home universe.
	if prev, ok, err := LoadRepoIndex(cacheRoot, UniverseHome); err != nil {
		return err
	} else if ok {
		for _, e := range prev.Repos {
			if !pathIsUnderRoot(absRoot, e.Path) {
				entries = append(entries, e)
			}
		}
	}

	seen := make(map[string]struct{}, len(entries)+len(repos))
	for _, e := range entries {
		seen[e.Path] = struct{}{}
	}
	for _, r := range repos {
		if r.Path == "" {
			continue
		}
		if _, dup := seen[r.Path]; dup {
			continue
		}
		seen[r.Path] = struct{}{}
		entries = append(entries, RepoIndexEntry{
			Path:     r.Path,
			RepoType: string(r.RepoType),
			GitDir:   r.GitDir,
			Depth:    depthFromRoot(absRoot, r.Path),
			SeenAt:   now,
		})
	}

	return SaveRepoIndex(cacheRoot, RepoIndex{
		Version:   1,
		Universe:  UniverseHome,
		Base:      absRoot,
		UpdatedAt: now,
		Repos:     entries,
	})
}

// pathIsUnderRoot reports whether path is absRoot or a descendant.
func pathIsUnderRoot(absRoot, path string) bool {
	absRoot = filepath.Clean(absRoot)
	path = filepath.Clean(path)
	if path == absRoot {
		return true
	}
	rel, err := filepath.Rel(absRoot, path)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// warmServeFromIndex serves live repos from a durable index under absRoot:
// ApplyLiveness, sibling ReadDir of unique parents, save-back, then build Repos.
//
// Sibling ReadDir is sync discovery work and is skipped when
// WalkConsumeSyncBudget(now − last_scan_end) is 0 (same tier table as walk-log
// consume). Missing last_scan_end is treated as ancient → full budget → sibling
// probe still runs (index-serve / default warm).
func warmServeFromIndex(ctx context.Context, absRoot, cacheRoot string, opts Options, idx RepoIndex, onRepo func(Repo) error) ([]Repo, warmServeStats, error) {
	start := time.Now()
	var stats warmServeStats

	// Restrict to this scan root, then drop dead .git paths.
	filtered := idx
	keptUnder := make([]RepoIndexEntry, 0, len(idx.Repos))
	for _, e := range idx.Repos {
		if pathIsUnderRoot(absRoot, e.Path) {
			keptUnder = append(keptUnder, e)
		}
	}
	filtered.Repos = keptUnder
	stats.candidates = len(filtered.Repos)

	liveIdx := ApplyLiveness(filtered)

	// Path set of live indexed repos (pre-sibling).
	livePaths := make(map[string]struct{}, len(liveIdx.Repos))
	entryByPath := make(map[string]RepoIndexEntry, len(liveIdx.Repos))
	for _, e := range liveIdx.Repos {
		livePaths[e.Path] = struct{}{}
		entryByPath[e.Path] = e
	}

	// Sibling probe shares the walk-consume sync budget tiers.
	nowT := resolveNow(opts)
	lastEnd := resolveLastScanEnd(opts, cacheRoot)
	delta := time.Hour
	if !lastEnd.IsZero() {
		delta = nowT.Sub(lastEnd)
	}
	syncBudget := WalkConsumeSyncBudget(delta)

	now := time.Now().UTC().Format(time.RFC3339)
	if syncBudget > 0 {
		// Sibling probe: for unique parents of live repos, ReadDir children with .git
		// that are not already indexed.
		siblings, err := discoverSiblingRepos(ctx, livePaths)
		if err != nil {
			stats.duration = time.Since(start)
			return nil, stats, err
		}
		for _, s := range siblings {
			if _, exists := livePaths[s.Path]; exists {
				continue
			}
			// Sibling ReadDir of parent may see peers outside absRoot
			// (e.g. Scan(A) when parent holds A+B); only keep under-root.
			if !pathIsUnderRoot(absRoot, s.Path) {
				continue
			}
			livePaths[s.Path] = struct{}{}
			entryByPath[s.Path] = RepoIndexEntry{
				Path:     s.Path,
				RepoType: string(s.RepoType),
				GitDir:   s.GitDir,
				Depth:    depthFromRoot(absRoot, s.Path),
				SeenAt:   now,
			}
		}
	}

	// Persist liveness + sibling discoveries under home universe.
	// Merge with index entries outside absRoot.
	saveEntries := make([]RepoIndexEntry, 0, len(entryByPath)+len(idx.Repos))
	if prev, ok, loadErr := LoadRepoIndex(cacheRoot, UniverseHome); loadErr != nil {
		stats.duration = time.Since(start)
		return nil, stats, loadErr
	} else if ok {
		for _, e := range prev.Repos {
			if !pathIsUnderRoot(absRoot, e.Path) {
				saveEntries = append(saveEntries, e)
			}
		}
	}
	for _, e := range entryByPath {
		saveEntries = append(saveEntries, e)
	}
	if err := SaveRepoIndex(cacheRoot, RepoIndex{
		Version:   1,
		Universe:  UniverseHome,
		Base:      absRoot,
		UpdatedAt: now,
		Repos:     saveEntries,
	}); err != nil {
		stats.duration = time.Since(start)
		return nil, stats, err
	}

	// Emit live repos (index + siblings) from live .git only (no mirror marks).
	var repos []Repo
	for path := range livePaths {
		select {
		case <-ctx.Done():
			stats.duration = time.Since(start)
			return nil, stats, ctx.Err()
		default:
		}

		repo, live, liveErr := liveRepoFromPath(path)
		if liveErr != nil {
			stats.duration = time.Since(start)
			return nil, stats, liveErr
		}
		if !live {
			continue
		}
		stats.live++
		if onRepo != nil {
			if err := onRepo(repo); err != nil {
				stats.duration = time.Since(start)
				return nil, stats, err
			}
		} else {
			repos = append(repos, repo)
		}
	}

	stats.duration = time.Since(start)
	return repos, stats, nil
}

// liveRepoFromPath builds a Repo when path/.git exists.
func liveRepoFromPath(path string) (Repo, bool, error) {
	gitPath := filepath.Join(path, ".git")
	info, statErr := os.Stat(gitPath)
	if statErr != nil || !(info.IsDir() || info.Mode().IsRegular()) {
		return Repo{}, false, nil
	}
	gitDir, repoType, resolveErr := resolveGitDir(path, gitPath, info)
	if resolveErr != nil {
		return Repo{
			Path:  path,
			Name:  filepath.Base(path),
			Error: resolveErr.Error(),
		}, true, nil
	}
	if gitDir == "" && repoType == "" {
		return Repo{}, false, nil
	}
	return Repo{
		Path:     path,
		Name:     filepath.Base(path),
		GitDir:   gitDir,
		RepoType: repoType,
	}, true, nil
}

// discoverSiblingRepos ReadDirs unique parents of live repo paths and returns
// child directories that have a .git marker and are not already in livePaths.
func discoverSiblingRepos(ctx context.Context, livePaths map[string]struct{}) ([]Repo, error) {
	parents := make(map[string]struct{})
	for p := range livePaths {
		parent := filepath.Dir(p)
		if parent == "" || parent == "." || parent == string(filepath.Separator) {
			continue
		}
		parents[parent] = struct{}{}
	}

	var found []Repo
	for parent := range parents {
		select {
		case <-ctx.Done():
			return found, ctx.Err()
		default:
		}
		entries, err := os.ReadDir(parent)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return found, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := e.Name()
			if name == "." || name == ".." || name == ".git" {
				continue
			}
			child := filepath.Join(parent, name)
			if _, exists := livePaths[child]; exists {
				continue
			}
			repo, live, err := liveRepoFromPath(child)
			if err != nil {
				return found, err
			}
			if !live {
				continue
			}
			found = append(found, repo)
		}
	}
	return found, nil
}
