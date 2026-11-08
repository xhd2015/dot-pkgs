package reclaim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectRepeat(t *testing.T) {
	t.Parallel()
	cases := []struct {
		path    string
		pattern string
		repeats int
	}{
		{"a/b/x/c/x/c/x/c/d", "x/c/", 3},
		{"/a/b/c", "", 0},
		{"/tmp/go-build1/b001/node_modules/pkg/node_modules/pkg/node_modules/pkg/x.js", "node_modules/pkg/", 3},
	}
	for _, c := range cases {
		p, n := detectRepeat(c.path)
		if p != c.pattern || n != c.repeats {
			t.Errorf("detectRepeat(%q) = (%q, %d), want (%q, %d)", c.path, p, n, c.pattern, c.repeats)
		}
	}
}

const logStuck = `Filtering the log data using "..."
{"timestamp":"2026-10-29 09:00:00.000000+0800","processID":11,"eventMessage":"CacheDeleteDaemonPurge purge_orphans urgency: 1, clearing: /private/var/dirs_cleaner/"}
{"timestamp":"2026-10-29 09:00:01.000000+0800","processID":11,"eventMessage":"nuke_dir: removefile error for /private/var/dirs_cleaner/ : File name too long"}
{"timestamp":"2026-10-29 10:00:00.000000+0800","processID":12,"eventMessage":"CacheDeleteDaemonPurge purge_orphans urgency: 1, clearing: /private/var/dirs_cleaner/"}
{"timestamp":"2026-10-29 10:00:01.000000+0800","processID":12,"eventMessage":"nuke_dir: removefile error for /private/var/dirs_cleaner/ : File name too long"}
{"timestamp":"2026-10-29 11:00:00.000000+0800","processID":13,"eventMessage":"CacheDeleteDaemonPurge purge_orphans urgency: 1, clearing: /private/var/dirs_cleaner/"}
{"timestamp":"2026-10-29 11:00:01.000000+0800","processID":13,"eventMessage":"nuke_dir: removefile error for /private/var/dirs_cleaner/ : File name too long"}
`

func TestParsePurgeLog(t *testing.T) {
	t.Parallel()
	ph := parsePurgeLog([]byte(logStuck))
	if ph.Attempts != 3 || ph.Failures != 3 || ph.LastOK || !ph.Stuck {
		t.Fatalf("stuck log: %+v", ph)
	}
	if ph.LastError != "File name too long" {
		t.Fatalf("last error = %q", ph.LastError)
	}

	ok := logStuck + `{"timestamp":"2026-10-29 12:00:00.000000+0800","processID":14,"eventMessage":"CacheDeleteDaemonPurge purge_orphans urgency: 1, clearing: /private/var/dirs_cleaner/"}` + "\n"
	ph = parsePurgeLog([]byte(ok))
	if ph.Attempts != 4 || ph.Failures != 3 || !ph.LastOK || ph.Stuck {
		t.Fatalf("recovered log: %+v", ph)
	}

	ph = parsePurgeLog([]byte("Filtering...\n"))
	if ph.Attempts != 0 || ph.Stuck {
		t.Fatalf("empty log: %+v", ph)
	}
}

func seedDirsCleaner(t *testing.T) string {
	t.Helper()
	root := filepath.Join(t.TempDir(), "dirs_cleaner")
	for _, batch := range []string{"aaaa", "bbbb"} {
		deep := filepath.Join(root, batch, "x", "c", "x", "c", "x", "c")
		if err := os.MkdirAll(deep, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(deep, "f"), make([]byte, 8192), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root
}

func TestAnalyseDirsCleaner_AccessibleRoot(t *testing.T) {
	t.Parallel()
	root := seedDirsCleaner(t)
	a, err := AnalyseDirsCleaner(Options{DirsCleanerRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	if len(a.Batches) != 2 || a.TotalBytes < 16384 {
		t.Fatalf("analysis: %+v", a)
	}
	for _, b := range a.Batches {
		if b.MaxPathLen < len(root)+len("/aaaa/x/c/x/c/x/c/f") {
			t.Fatalf("batch %s max path %d too short", b.Name, b.MaxPathLen)
		}
	}
}

func TestPurgeDirsCleaner_KeepsRoot(t *testing.T) {
	t.Parallel()
	root := seedDirsCleaner(t)
	res, err := PurgeDirsCleaner(Options{DirsCleanerRoot: root})
	if err != nil {
		t.Fatal(err)
	}
	ents, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("root must survive: %v", err)
	}
	if len(ents) != 0 || res.Left != 0 || res.Before == 0 {
		t.Fatalf("entries=%d outcome=%+v", len(ents), res)
	}
}

func lockedRoot(t *testing.T) string {
	t.Helper()
	root := seedDirsCleaner(t)
	if err := os.Chmod(root, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(root, 0o755) })
	if _, err := os.ReadDir(root); err == nil {
		t.Skip("running as root; cannot simulate an unreadable dirs_cleaner")
	}
	return root
}

func TestPurgeDirsCleaner_NeedsAdminWithoutTTY(t *testing.T) {
	t.Parallel()
	root := lockedRoot(t)
	_, err := PurgeDirsCleaner(Options{
		DirsCleanerRoot: root,
		Geteuid:         func() int { return 501 },
		IsTerminal:      func() bool { return false },
		ElevateMode:     ElevateSudo,
	})
	if err == nil || !strings.Contains(err.Error(), "requires admin privileges") {
		t.Fatalf("err = %v", err)
	}
}

func TestPurgeDirsCleaner_InjectedElevateReportsRestricted(t *testing.T) {
	t.Parallel()
	root := lockedRoot(t)
	var gotPurpose, gotScript string
	res, err := PurgeDirsCleaner(Options{
		DirsCleanerRoot: root,
		Geteuid:         func() int { return 501 },
		Elevate: func(purpose, script string) ([]byte, error) {
			gotPurpose, gotScript = purpose, script
			return []byte("BEFORE\t2048\nAFTER\t4\nLEFT\t1\nR\taaaa/locked.pkg\n"), nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if gotPurpose != "dirs-cleaner purge" || !strings.Contains(gotScript, shQuote(root)) {
		t.Fatalf("purpose=%q script=%q", gotPurpose, gotScript)
	}
	if res.Before != 2048*1024 || res.Left != 1 || len(res.Restricted) != 1 || res.Restricted[0] != "aaaa/locked.pkg" {
		t.Fatalf("outcome: %+v", res)
	}
}

func TestOsascriptRefusesNonDefaultRoot(t *testing.T) {
	t.Parallel()
	root := lockedRoot(t)
	_, err := AnalyseDirsCleaner(Options{DirsCleanerRoot: root, Geteuid: func() int { return 501 }, ElevateMode: ElevateOSAScript})
	if err == nil || !strings.Contains(err.Error(), "refusing to elevate") {
		t.Fatalf("err = %v", err)
	}
}
