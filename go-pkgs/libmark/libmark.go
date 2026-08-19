// Package libmark persists mark(1) records under ~/.mark and resolves them by PID.
package libmark

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/proc"
)

const (
	runningDir = "running"
	archiveDir = "archived"
)

// Record is one mark process, live or archived.
type Record struct {
	PID       int        `json:"pid"`
	Content   string     `json:"content"`
	Dir       string     `json:"dir"`
	CreatedAt time.Time  `json:"created_at"`
	ExitedAt  *time.Time `json:"exited_at,omitempty"`
	Git       *GitInfo   `json:"git,omitempty"`
}

// GitInfo is a snapshot of the git checkout at mark start. Omitted when dir is not a repo.
type GitInfo struct {
	Toplevel string `json:"toplevel"`
	GitDir   string `json:"gitdir"`
	Commit   string `json:"commit,omitempty"`
	Branch   string `json:"branch,omitempty"`
	Origin   string `json:"origin,omitempty"`
}

// Store reads and writes a mark root (default ~/.mark). Hooks are for tests.
type Store struct {
	Root    string
	Now     func() time.Time
	Alive   func(pid int) bool
	LookGit func(dir string) *GitInfo
	Getwd   func() (string, error)
}

// DefaultRoot is $HOME/.mark.
func DefaultRoot() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return ".mark"
	}
	return filepath.Join(home, ".mark")
}

// Default uses ~/.mark and live process/git probes.
func Default() *Store {
	return &Store{Root: DefaultRoot()}
}

// Resolve reads the live record for pid. A dead PID is archived first, then missed.
func Resolve(pid int) (*Record, error) {
	return Default().Resolve(pid)
}

// WriteLive writes a running record under the default root.
func WriteLive(rec Record) error {
	return Default().WriteLive(rec)
}

// Archive moves a live record to archived/ under the default root.
func Archive(pid int) error {
	return Default().Archive(pid)
}

// Sweep archives every live file whose PID is no longer running.
func Sweep() error {
	return Default().Sweep()
}

func (s *Store) root() string {
	if s == nil || strings.TrimSpace(s.Root) == "" {
		return DefaultRoot()
	}
	return s.Root
}

func (s *Store) runningDir() string {
	return filepath.Join(s.root(), runningDir)
}

func (s *Store) archiveDir() string {
	return filepath.Join(s.root(), archiveDir)
}

// RunningPath is ~/.mark/running/<pid>.json.
func (s *Store) RunningPath(pid int) string {
	return filepath.Join(s.runningDir(), fmt.Sprintf("%d.json", pid))
}

// ArchivePath is ~/.mark/archived/<pid>-<YYYYMMDDThhmmss±ZZZZ>.json.
func (s *Store) ArchivePath(rec Record) string {
	return filepath.Join(s.archiveDir(), archiveFileName(rec.PID, rec.CreatedAt))
}

func archiveFileName(pid int, created time.Time) string {
	return fmt.Sprintf("%d-%s.json", pid, created.Format("20060102T150405-0700"))
}

func (s *Store) now() time.Time {
	if s != nil && s.Now != nil {
		return s.Now().Truncate(time.Second)
	}
	return time.Now().Truncate(time.Second)
}

func (s *Store) alive(pid int) bool {
	if s != nil && s.Alive != nil {
		return s.Alive(pid)
	}
	return proc.Alive(pid, proc.Options{})
}

func (s *Store) getwd() (string, error) {
	if s != nil && s.Getwd != nil {
		return s.Getwd()
	}
	return os.Getwd()
}

func (s *Store) lookGit(dir string) *GitInfo {
	if s != nil && s.LookGit != nil {
		return s.LookGit(dir)
	}
	return CaptureGit(dir)
}

// WriteLive fills missing pid/dir/time/git and writes running/<pid>.json.
func (s *Store) WriteLive(rec Record) error {
	if rec.PID == 0 {
		rec.PID = os.Getpid()
	}
	if rec.PID <= 0 {
		return fmt.Errorf("libmark: invalid pid %d", rec.PID)
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = s.now()
	} else {
		rec.CreatedAt = rec.CreatedAt.Truncate(time.Second)
	}
	if strings.TrimSpace(rec.Dir) == "" {
		dir, err := s.getwd()
		if err != nil {
			return fmt.Errorf("libmark: getwd: %w", err)
		}
		rec.Dir = dir
	}
	if rec.Git == nil {
		rec.Git = s.lookGit(rec.Dir)
	}
	rec.ExitedAt = nil
	if err := os.MkdirAll(s.runningDir(), 0o755); err != nil {
		return fmt.Errorf("libmark: mkdir running: %w", err)
	}
	if err := writeJSON(s.RunningPath(rec.PID), rec); err != nil {
		return fmt.Errorf("libmark: write live: %w", err)
	}
	return nil
}

// Resolve returns the live record for pid. Dead PIDs are archived, then missed.
func (s *Store) Resolve(pid int) (*Record, error) {
	rec, err := s.readRunning(pid)
	if err != nil {
		return nil, err
	}
	if !s.alive(pid) {
		if aerr := s.archiveRecord(*rec); aerr != nil {
			return nil, aerr
		}
		return nil, fmt.Errorf("libmark: pid %d is not running", pid)
	}
	return rec, nil
}

// Archive moves running/<pid>.json to archived/. Missing live files are a no-op.
func (s *Store) Archive(pid int) error {
	rec, err := s.readRunning(pid)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return s.archiveRecord(*rec)
}

// Sweep archives live files whose PID is dead.
func (s *Store) Sweep() error {
	ents, err := os.ReadDir(s.runningDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("libmark: read running: %w", err)
	}
	var first error
	for _, ent := range ents {
		if ent.IsDir() {
			continue
		}
		pid, ok := parseRunningName(ent.Name())
		if !ok {
			continue
		}
		if s.alive(pid) {
			continue
		}
		if err := s.Archive(pid); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func parseRunningName(name string) (int, bool) {
	if !strings.HasSuffix(name, ".json") {
		return 0, false
	}
	id := strings.TrimSuffix(name, ".json")
	pid, err := strconv.Atoi(id)
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

func (s *Store) readRunning(pid int) (*Record, error) {
	if pid <= 0 {
		return nil, fmt.Errorf("libmark: invalid pid %d", pid)
	}
	return readJSON(s.RunningPath(pid))
}

func (s *Store) archiveRecord(rec Record) error {
	now := s.now()
	rec.ExitedAt = &now
	if err := os.MkdirAll(s.archiveDir(), 0o755); err != nil {
		return fmt.Errorf("libmark: mkdir archived: %w", err)
	}
	if err := writeJSON(s.ArchivePath(rec), rec); err != nil {
		return fmt.Errorf("libmark: write archive: %w", err)
	}
	if err := os.Remove(s.RunningPath(rec.PID)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("libmark: remove live: %w", err)
	}
	return nil
}

func writeJSON(path string, rec Record) error {
	data, err := json.MarshalIndent(rec, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func readJSON(path string) (*Record, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("libmark: read %s: %w", path, err)
	}
	var rec Record
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("libmark: parse %s: %w", path, err)
	}
	return &rec, nil
}
