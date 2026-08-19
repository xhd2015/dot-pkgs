package libmark

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	loc := time.FixedZone("CST", 8*3600)
	created := time.Date(2026, 8, 18, 4, 7, 42, 0, loc)
	return &Store{
		Root:    t.TempDir(),
		Now:     func() time.Time { return created },
		Alive:   func(int) bool { return true },
		LookGit: func(string) *GitInfo { return nil },
		Getwd:   func() (string, error) { return "/tmp/mark-wd", nil },
	}
}

func TestWriteLiveResolve(t *testing.T) {
	s := testStore(t)
	err := s.WriteLive(Record{PID: 14460, Content: "I'm still waiting for result"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Resolve(14460)
	if err != nil {
		t.Fatal(err)
	}
	if got.PID != 14460 {
		t.Fatalf("pid: got %d", got.PID)
	}
	if got.Content != "I'm still waiting for result" {
		t.Fatalf("content: %q", got.Content)
	}
	if got.Dir != "/tmp/mark-wd" {
		t.Fatalf("dir: %q", got.Dir)
	}
	if got.Git != nil {
		t.Fatalf("git: %+v", got.Git)
	}
	if got.ExitedAt != nil {
		t.Fatalf("exited_at set on live: %v", got.ExitedAt)
	}
	want := time.Date(2026, 8, 18, 4, 7, 42, 0, time.FixedZone("CST", 8*3600))
	if !got.CreatedAt.Equal(want) {
		t.Fatalf("created_at: %v", got.CreatedAt)
	}
	raw, err := os.ReadFile(s.RunningPath(14460))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"created_at": "2026-08-18T04:07:42+08:00"`) {
		t.Fatalf("json missing offset created_at:\n%s", raw)
	}
	if !strings.Contains(string(raw), `"dir": "/tmp/mark-wd"`) {
		t.Fatalf("json missing dir:\n%s", raw)
	}
}

func TestEmptyContent(t *testing.T) {
	s := testStore(t)
	if err := s.WriteLive(Record{PID: 1, Content: ""}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Resolve(1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Content != "" {
		t.Fatalf("content: %q", got.Content)
	}
}

func TestArchiveMovesAndResolveMisses(t *testing.T) {
	s := testStore(t)
	if err := s.WriteLive(Record{PID: 99, Content: "bye"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Archive(99); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.RunningPath(99)); !os.IsNotExist(err) {
		t.Fatalf("live still exists: %v", err)
	}
	arch := filepath.Join(s.archiveDir(), "99-20260818T040742+0800.json")
	raw, err := os.ReadFile(arch)
	if err != nil {
		t.Fatal(err)
	}
	var rec Record
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.Content != "bye" || rec.ExitedAt == nil {
		t.Fatalf("archived: %+v", rec)
	}
	if rec.ExitedAt.Format(time.RFC3339) != "2026-08-18T04:07:42+08:00" {
		t.Fatalf("exited_at: %v", rec.ExitedAt)
	}
	if _, err := s.Resolve(99); err == nil {
		t.Fatal("Resolve after Archive: want error")
	}
}

func TestArchiveMissingIsNoop(t *testing.T) {
	s := testStore(t)
	if err := s.Archive(404); err != nil {
		t.Fatal(err)
	}
}

func TestSweepDeadArchives(t *testing.T) {
	s := testStore(t)
	s.Alive = func(pid int) bool { return pid == 2 }
	if err := s.WriteLive(Record{PID: 1, Content: "dead"}); err != nil {
		t.Fatal(err)
	}
	if err := s.WriteLive(Record{PID: 2, Content: "live"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Sweep(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(s.RunningPath(1)); !os.IsNotExist(err) {
		t.Fatal("dead pid still live")
	}
	if _, err := s.Resolve(2); err != nil {
		t.Fatalf("live pid swept: %v", err)
	}
	if _, err := os.Stat(filepath.Join(s.archiveDir(), "1-20260818T040742+0800.json")); err != nil {
		t.Fatal(err)
	}
}

func TestResolveDeadArchives(t *testing.T) {
	s := testStore(t)
	s.Alive = func(int) bool { return false }
	if err := s.WriteLive(Record{PID: 7, Content: "gone"}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Resolve(7); err == nil {
		t.Fatal("Resolve dead pid: want error")
	}
	if _, err := os.Stat(s.RunningPath(7)); !os.IsNotExist(err) {
		t.Fatal("dead pid still live after Resolve")
	}
}

func TestTwoArchivesSamePID(t *testing.T) {
	s := testStore(t)
	loc := time.FixedZone("CST", 8*3600)
	t1 := time.Date(2026, 8, 18, 4, 7, 42, 0, loc)
	t2 := time.Date(2026, 8, 18, 5, 0, 0, 0, loc)
	s.Now = func() time.Time { return t1 }
	if err := s.WriteLive(Record{PID: 8, Content: "first"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Archive(8); err != nil {
		t.Fatal(err)
	}
	s.Now = func() time.Time { return t2 }
	if err := s.WriteLive(Record{PID: 8, Content: "second"}); err != nil {
		t.Fatal(err)
	}
	if err := s.Archive(8); err != nil {
		t.Fatal(err)
	}
	a1 := filepath.Join(s.archiveDir(), "8-20260818T040742+0800.json")
	a2 := filepath.Join(s.archiveDir(), "8-20260818T050000+0800.json")
	if _, err := os.Stat(a1); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(a2); err != nil {
		t.Fatal(err)
	}
}

func TestWriteLiveKeepsGit(t *testing.T) {
	s := testStore(t)
	s.LookGit = func(dir string) *GitInfo {
		if dir != "/repo" {
			t.Fatalf("LookGit dir: %q", dir)
		}
		return &GitInfo{
			Toplevel: "/repo",
			GitDir:   "/repo/.git",
			Commit:   "abc",
			Branch:   "main",
			Origin:   "git@example.com:x/y.git",
		}
	}
	if err := s.WriteLive(Record{PID: 3, Content: "g", Dir: "/repo"}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Resolve(3)
	if err != nil {
		t.Fatal(err)
	}
	if got.Git == nil || got.Git.Toplevel != "/repo" || got.Git.GitDir != "/repo/.git" || got.Git.Commit != "abc" || got.Git.Branch != "main" || got.Git.Origin != "git@example.com:x/y.git" {
		t.Fatalf("git: %+v", got.Git)
	}
}

func TestCaptureGitNotRepo(t *testing.T) {
	if got := CaptureGit(t.TempDir()); got != nil {
		t.Fatalf("want nil, got %+v", got)
	}
}

func TestCaptureGitRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "t@t")
	run("config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "f")
	run("commit", "-m", "init")
	run("remote", "add", "origin", "git@example.com:x/y.git")

	got := CaptureGit(dir)
	if got == nil {
		t.Fatal("CaptureGit: nil")
	}
	if got.Toplevel != dir {
		// macOS /var -> /private/var
		if real, err := filepath.EvalSymlinks(dir); err != nil || got.Toplevel != real {
			t.Fatalf("toplevel: %q dir=%q", got.Toplevel, dir)
		}
	}
	if !strings.HasSuffix(got.GitDir, ".git") {
		t.Fatalf("gitdir: %q", got.GitDir)
	}
	if len(got.Commit) != 40 {
		t.Fatalf("commit: %q", got.Commit)
	}
	if got.Branch != "main" {
		t.Fatalf("branch: %q", got.Branch)
	}
	if got.Origin != "git@example.com:x/y.git" {
		t.Fatalf("origin: %q", got.Origin)
	}
}
