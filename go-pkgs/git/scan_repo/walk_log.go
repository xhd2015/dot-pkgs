package scan_repo

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// WalkLogEvent is one append-only JSONL record under home/walk.jsonl.
type WalkLogEvent struct {
	Op   string `json:"op"`
	Path string `json:"path,omitempty"`
	Gen  int    `json:"gen,omitempty"`
}

// WalkCursor is the durable byte offset into walk.jsonl.
type WalkCursor struct {
	Offset int64 `json:"offset"`
}

// homeMeta is optional durable metadata under home/meta.json.
type homeMeta struct {
	LastScanEnd string `json:"last_scan_end,omitempty"`
}

// WalkLogPath returns <cacheRoot>/home/walk.jsonl.
func WalkLogPath(cacheRoot string) string {
	return filepath.Join(cacheRoot, UniverseHome, "walk.jsonl")
}

// WalkCursorPath returns <cacheRoot>/home/walk.cursor.json.
func WalkCursorPath(cacheRoot string) string {
	return filepath.Join(cacheRoot, UniverseHome, "walk.cursor.json")
}

// HomeMetaPath returns <cacheRoot>/home/meta.json.
func HomeMetaPath(cacheRoot string) string {
	return filepath.Join(cacheRoot, UniverseHome, "meta.json")
}

// WalkConsumeSyncBudget returns the wall-clock budget for sync walk-log consume
// given how long has passed since last_scan_end:
//
//	delta < 10s        → 0 (side / best-effort only; no sync re-list)
//	10s ≤ delta < 60s  → 500ms
//	delta ≥ 60s        → 1s
func WalkConsumeSyncBudget(sinceLast time.Duration) time.Duration {
	if sinceLast < 10*time.Second {
		return 0
	}
	if sinceLast < 60*time.Second {
		return 500 * time.Millisecond
	}
	return time.Second
}

// WalkLogAppend appends one JSON object as a line to home/walk.jsonl.
// No-op when cacheRoot is empty.
func WalkLogAppend(cacheRoot string, ev WalkLogEvent) error {
	if cacheRoot == "" {
		return nil
	}
	path := WalkLogPath(cacheRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(append(data, '\n')); err != nil {
		return err
	}
	return nil
}

// AppendWalkVisit records a directory visit during cold walk.
func AppendWalkVisit(cacheRoot, dirPath string) error {
	if cacheRoot == "" {
		return nil
	}
	return WalkLogAppend(cacheRoot, WalkLogEvent{
		Op:   "visit",
		Path: filepath.Clean(dirPath),
	})
}

// SaveWalkCursor writes home/walk.cursor.json with the given byte offset.
// No-op when cacheRoot is empty.
func SaveWalkCursor(cacheRoot string, offset int64) error {
	if cacheRoot == "" {
		return nil
	}
	path := WalkCursorPath(cacheRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.Marshal(WalkCursor{Offset: offset})
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// LoadWalkCursor reads home/walk.cursor.json. ok is false when missing.
func LoadWalkCursor(cacheRoot string) (WalkCursor, bool, error) {
	if cacheRoot == "" {
		return WalkCursor{}, false, nil
	}
	raw, err := os.ReadFile(WalkCursorPath(cacheRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return WalkCursor{}, false, nil
		}
		return WalkCursor{}, false, err
	}
	var cur WalkCursor
	if err := json.Unmarshal(raw, &cur); err != nil {
		return WalkCursor{}, false, err
	}
	return cur, true, nil
}

// SealColdWalkGenEnd appends {"op":"gen_end","gen":gen} and sets the walk
// cursor offset to the sealed walk.jsonl byte length. First cold uses gen=1.
// No-op when cacheRoot is empty.
func SealColdWalkGenEnd(cacheRoot string, gen int) error {
	if cacheRoot == "" {
		return nil
	}
	if err := WalkLogAppend(cacheRoot, WalkLogEvent{Op: "gen_end", Gen: gen}); err != nil {
		return err
	}
	st, err := os.Stat(WalkLogPath(cacheRoot))
	if err != nil {
		return err
	}
	return SaveWalkCursor(cacheRoot, st.Size())
}

// SaveLastScanEnd persists last_scan_end (RFC3339) under home/meta.json.
// No-op when cacheRoot is empty.
func SaveLastScanEnd(cacheRoot string, t time.Time) error {
	if cacheRoot == "" || t.IsZero() {
		return nil
	}
	path := HomeMetaPath(cacheRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	meta := homeMeta{LastScanEnd: t.UTC().Format(time.RFC3339Nano)}
	// Merge with existing meta if present.
	if raw, err := os.ReadFile(path); err == nil {
		var prev homeMeta
		if json.Unmarshal(raw, &prev) == nil && prev.LastScanEnd != "" && meta.LastScanEnd == "" {
			meta = prev
		}
	}
	data, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	tmp := fmt.Sprintf("%s.tmp.%d", path, os.Getpid())
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

// LoadLastScanEnd reads last_scan_end from home/meta.json. ok is false when
// missing or unparseable.
func LoadLastScanEnd(cacheRoot string) (time.Time, bool) {
	if cacheRoot == "" {
		return time.Time{}, false
	}
	raw, err := os.ReadFile(HomeMetaPath(cacheRoot))
	if err != nil {
		return time.Time{}, false
	}
	var meta homeMeta
	if err := json.Unmarshal(raw, &meta); err != nil || meta.LastScanEnd == "" {
		return time.Time{}, false
	}
	// Prefer RFC3339 / RFC3339Nano; also accept unix seconds as decimal string.
	if t, err := time.Parse(time.RFC3339Nano, meta.LastScanEnd); err == nil {
		return t, true
	}
	if t, err := time.Parse(time.RFC3339, meta.LastScanEnd); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// resolveLastScanEnd prefers Options.LastScanEnd, then meta.json. Zero means
// "ancient" (full sync budget).
func resolveLastScanEnd(opts Options, cacheRoot string) time.Time {
	if !opts.LastScanEnd.IsZero() {
		return opts.LastScanEnd
	}
	if t, ok := LoadLastScanEnd(cacheRoot); ok {
		return t
	}
	return time.Time{}
}

// readWalkLogEvents parses non-empty JSONL lines from walk.jsonl.
func readWalkLogEvents(cacheRoot string) ([]WalkLogEvent, error) {
	path := WalkLogPath(cacheRoot)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var events []WalkLogEvent
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := bytes.TrimSpace(sc.Bytes())
		if len(line) == 0 {
			continue
		}
		var ev WalkLogEvent
		if err := json.Unmarshal(line, &ev); err != nil {
			return nil, err
		}
		events = append(events, ev)
	}
	return events, sc.Err()
}

// consumeWalkLog re-lists visit paths from the last sealed generation under a
// sync budget, appends gone/new events at EOF, seals gen_end G+1 when gen_end G
// is fully processed, and advances the walk cursor to the new log EOF.
//
// Budget 0: no sync work (cursor / log unchanged for consume purposes).
// New repos discovered during re-list are merged into the returned slice (and
// via onRepo when non-nil).
func consumeWalkLog(ctx context.Context, cacheRoot string, opts Options, absRoot string, existing []Repo, onRepo func(Repo) error) ([]Repo, error) {
	if cacheRoot == "" {
		return existing, nil
	}
	logPath := WalkLogPath(cacheRoot)
	if _, err := os.Stat(logPath); err != nil {
		return existing, nil
	}

	now := resolveNow(opts)
	lastEnd := resolveLastScanEnd(opts, cacheRoot)
	var delta time.Duration
	if lastEnd.IsZero() {
		// Missing last_scan_end → treat as ancient / full budget.
		delta = time.Hour
	} else {
		delta = now.Sub(lastEnd)
	}
	budget := WalkConsumeSyncBudget(delta)
	if budget <= 0 {
		// Still stamp meta so subsequent scans can age correctly.
		_ = SaveLastScanEnd(cacheRoot, now)
		return existing, nil
	}

	events, err := readWalkLogEvents(cacheRoot)
	if err != nil {
		return existing, err
	}
	if len(events) == 0 {
		return existing, nil
	}

	// Last sealed generation G.
	lastGen := 0
	lastGenIdx := -1
	for i, ev := range events {
		if ev.Op == "gen_end" && ev.Gen > 0 {
			lastGen = ev.Gen
			lastGenIdx = i
		}
	}
	if lastGenIdx < 0 || lastGen < 1 {
		return existing, nil
	}

	// Visits belonging to generation G: after previous gen_end through last gen_end.
	prevGenEndIdx := -1
	for i := lastGenIdx - 1; i >= 0; i-- {
		if events[i].Op == "gen_end" {
			prevGenEndIdx = i
			break
		}
	}
	var visits []string
	seenVisit := make(map[string]struct{})
	for i := prevGenEndIdx + 1; i < lastGenIdx; i++ {
		ev := events[i]
		if ev.Op != "visit" || ev.Path == "" {
			continue
		}
		p := filepath.Clean(ev.Path)
		if _, ok := seenVisit[p]; ok {
			continue
		}
		seenVisit[p] = struct{}{}
		visits = append(visits, p)
	}

	// Known paths already in the log (any op with path).
	known := make(map[string]struct{}, len(events))
	for _, ev := range events {
		if ev.Path != "" {
			known[filepath.Clean(ev.Path)] = struct{}{}
		}
	}

	// Dedupe repos already served / known.
	seenRepo := make(map[string]struct{}, len(existing))
	for _, r := range existing {
		if r.Path != "" {
			seenRepo[filepath.Clean(r.Path)] = struct{}{}
		}
	}

	deadline := time.Now().Add(budget)
	repos := existing
	finished := true

	addRepo := func(repo Repo) error {
		p := filepath.Clean(repo.Path)
		if _, ok := seenRepo[p]; ok {
			return nil
		}
		seenRepo[p] = struct{}{}
		// Dense mirror retired: discoveries merge into index after consume.
		if onRepo != nil {
			return onRepo(repo)
		}
		repos = append(repos, repo)
		return nil
	}

	for _, vp := range visits {
		if time.Now().After(deadline) {
			finished = false
			break
		}
		select {
		case <-ctx.Done():
			return repos, ctx.Err()
		default:
		}

		st, statErr := os.Stat(vp)
		if statErr != nil || !st.IsDir() {
			if err := WalkLogAppend(cacheRoot, WalkLogEvent{Op: "gone", Path: vp}); err != nil {
				return repos, err
			}
			continue
		}

		// Re-list direct children for new dirs / repos.
		entries, rdErr := os.ReadDir(vp)
		if rdErr != nil {
			continue
		}
		for _, e := range entries {
			if time.Now().After(deadline) {
				finished = false
				break
			}
			if !e.IsDir() || e.Name() == ".git" {
				continue
			}
			child := filepath.Clean(filepath.Join(vp, e.Name()))
			if _, ok := known[child]; !ok {
				if err := WalkLogAppend(cacheRoot, WalkLogEvent{Op: "visit", Path: child}); err != nil {
					return repos, err
				}
				known[child] = struct{}{}
			}
			// Discover a checkout at this child.
			if repo, ok := liveRepoAt(child); ok {
				if err := addRepo(repo); err != nil {
					return repos, err
				}
			}
		}
		if !finished {
			break
		}
	}

	// When gen_end G is fully consumed, seal gen_end G+1.
	if finished {
		if err := WalkLogAppend(cacheRoot, WalkLogEvent{Op: "gen_end", Gen: lastGen + 1}); err != nil {
			return repos, err
		}
	}

	// Advance cursor to sealed (or partial-append) EOF.
	st, err := os.Stat(logPath)
	if err != nil {
		return repos, err
	}
	if err := SaveWalkCursor(cacheRoot, st.Size()); err != nil {
		return repos, err
	}

	// Merge any newly discovered repos into durable index (best-effort).
	if absRoot != "" {
		// Only seed entries we added beyond existing; seedHomeRepoIndex merges.
		var newly []Repo
		for _, r := range repos {
			p := filepath.Clean(r.Path)
			found := false
			for _, e := range existing {
				if filepath.Clean(e.Path) == p {
					found = true
					break
				}
			}
			if !found {
				newly = append(newly, r)
			}
		}
		if len(newly) > 0 {
			_ = seedHomeRepoIndex(cacheRoot, absRoot, append(append([]Repo{}, existing...), newly...))
		}
	}

	_ = SaveLastScanEnd(cacheRoot, now)
	return repos, nil
}

// liveRepoAt classifies path as a live git checkout when path/.git exists.
func liveRepoAt(path string) (Repo, bool) {
	gitPath := filepath.Join(path, ".git")
	info, err := os.Stat(gitPath)
	if err != nil || !(info.IsDir() || info.Mode().IsRegular()) {
		return Repo{}, false
	}
	gitDir, repoType, resolveErr := resolveGitDir(path, gitPath, info)
	if resolveErr != nil {
		return Repo{
			Path:  path,
			Name:  filepath.Base(path),
			Error: resolveErr.Error(),
		}, true
	}
	if gitDir == "" && repoType == "" {
		return Repo{}, false
	}
	return Repo{
		Path:     path,
		Name:     filepath.Base(path),
		GitDir:   gitDir,
		RepoType: repoType,
	}, true
}
