# Scenario

**Feature**: headless geometry for tui/mouse via fixture-inline + **tty-watch binary**

```
# no Go module dep on tty-watch — binary only (breaks cycle with ptywrap)
Setup -> go build fixture-inline; resolve tty-watch bin (PATH / build sibling)
Harness -> exec tty-watch run --detach -- fixture --anchor
Fixture paint + CSI 6n -> CPR -> ORIGIN=<n> VIEW=<v>
Harness -> exec tty-watch send/snapshot/kill
```

## Preconditions

- Go toolchain on PATH.
- Module root is `DOCTEST_ROOT/../../../..` from this nested tree
  (`headless` → `tests` → `mouse` → `tui` → go-pkgs).
- **tty-watch binary** (not a go.mod require): look up order in `ensureBins`:
  1. `TTY_WATCH_BIN` env
  2. `exec.LookPath("tty-watch")`
  3. `go build` from `TTY_WATCH_DIR` or sibling brought tree
     `moduleRoot/../../tty-watch-master-2026-07-19`
- go-pkgs **must not** import or replace `github.com/xhd2015/tty-watch`.
- Fixture package: `tui/mouse/cmd/fixture-inline`.
- Shared only across leaves: compiled binaries under
  `$TMPDIR/mouse-headless-doctest-<DOCTEST_SESSION_ID>/`.
- Per-leaf: temp Home dir and unique SessionID (mutated state never shared).

## Context

- Nested `DOCTEST.md` firewall: pure tree Run/Request do not apply here.
- Labels `needs-pty,slow` on all leaves.
- Host is always **subprocess** (`exec.Command`), never `cli.Run` import.

## Steps

1. Ensure session binary cache (fixture + tty-watch) under flock.
2. Allocate per-leaf Home + SessionID.
3. Branch Setup sets Anchor; leaf Setup sets Action / click fields.
4. Root Run drives detach → poll → send → kill.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"
	"github.com/xhd2015/doctest/session"
)

func sessionCacheDir(d *session.Doctest) string {
	return filepath.Join(os.TempDir(), "mouse-headless-doctest-"+d.DOCTEST_SESSION_ID)
}

func moduleRoot(d *session.Doctest) string {
	// headless -> tests -> mouse -> tui -> go-pkgs
	return filepath.Clean(filepath.Join(d.DOCTEST_ROOT, "..", "..", "..", ".."))
}

func ttyWatchSrcDir(d *session.Doctest) string {
	if d := os.Getenv("TTY_WATCH_DIR"); d != "" {
		return d
	}
	// go-pkgs is external/dot-pkgs-.../go-pkgs; tty-watch is sibling external/tty-watch-...
	return filepath.Clean(filepath.Join(moduleRoot(d), "..", "..", "tty-watch-master-2026-07-19"))
}

func withFileLock(lockPath string, fn func() error) error {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
	return fn()
}

func fileExists(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// resolveTTYWatchBin: TTY_WATCH_BIN → PATH → go build from TTY_WATCH_DIR / sibling tree.
// Never imports the tty-watch Go module (binary dependency only).
func resolveTTYWatchBin(d *session.Doctest, cache string) (string, error) {
	if p := os.Getenv("TTY_WATCH_BIN"); p != "" {
		if fileExists(p) {
			return p, nil
		}
		return "", fmt.Errorf("TTY_WATCH_BIN=%q not found", p)
	}
	if p, err := exec.LookPath("tty-watch"); err == nil && p != "" {
		return p, nil
	}
	outBin := filepath.Join(cache, "tty-watch")
	if fileExists(outBin) {
		return outBin, nil
	}
	twSrc := ttyWatchSrcDir(d)
	if st, err := os.Stat(filepath.Join(twSrc, "cmd", "tty-watch")); err != nil || !st.IsDir() {
		return "", fmt.Errorf("tty-watch not on PATH and no source at %s (set TTY_WATCH_BIN or TTY_WATCH_DIR)", twSrc)
	}
	cmd := exec.Command("go", "build", "-o", outBin, "./cmd/tty-watch")
	cmd.Dir = twSrc
	if out, e := cmd.CombinedOutput(); e != nil {
		return "", fmt.Errorf("build tty-watch from %s: %w\n%s", twSrc, e, out)
	}
	return outBin, nil
}

func ensureBins(t *testing.T, d *session.Doctest) (fixtureBin, ttyWatchBin string, err error) {
	t.Helper()
	cache := sessionCacheDir(d)
	if err := os.MkdirAll(cache, 0o755); err != nil {
		return "", "", err
	}
	lock := filepath.Join(cache, "build.lock")
	ready := filepath.Join(cache, "binaries.ready")
	fixtureBin = filepath.Join(cache, "fixture-inline")

	err = withFileLock(lock, func() error {
		if !fileExists(fixtureBin) {
			mod := moduleRoot(d)
			cmd := exec.Command("go", "build", "-o", fixtureBin, "./tui/mouse/cmd/fixture-inline")
			cmd.Dir = mod
			if out, e := cmd.CombinedOutput(); e != nil {
				return fmt.Errorf("build fixture-inline: %w\n%s", e, out)
			}
		}
		tw, e := resolveTTYWatchBin(d, cache)
		if e != nil {
			return e
		}
		ttyWatchBin = tw
		if !fileExists(ready) {
			_ = os.WriteFile(ready, []byte("ok"), 0o644)
		}
		return nil
	})
	return fixtureBin, ttyWatchBin, err
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	t.Helper()
	if req == nil {
		return fmt.Errorf("nil request")
	}
	fixBin, twBin, err := ensureBins(t, d)
	if err != nil {
		return err
	}
	req.FixtureBin = fixBin
	req.TTYWatchBin = twBin
	req.Home = t.TempDir()
	req.SessionID = "mhl-" + strconv.FormatInt(time.Now().UnixNano(), 36)
	if req.ClickCol <= 0 {
		req.ClickCol = 5
	}
	if req.Anchor == "" {
		req.Anchor = "mid"
	}
	return nil
}
```
