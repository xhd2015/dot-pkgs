# wrk bash-integration follow-up auto-cd

## Version
0.0.4

Decision tree covering bash follow-up auto-`cd`: the binary writes
`cd /abs` lines to `WRK_FOLLOWUP_FILE` after successful create / `--done` /
`--set-task` move (with a **home gate** for create and a **shell-cwd
existence gate** for done/set-task); the installed `wrk()` bash wrapper
executes whitelisted follow-ups; `--no-cd` and `WRK_AUTO_CD=0` suppress
the protocol.

# DSN (Domain Specific Notion)

- **wrk binary** — session-built CLI. When `WRK_FOLLOWUP_FILE` is set and
  `--no-cd` was not passed, on success of create / `--done` (remove) /
  `--set-task` (path actually moved) it may append one line
  `cd /absolute/path` to that file. Unset env → zero follow-up side effects.
  Failures, aborted confirms, `--merge-back`, inspect modes, and
  `--set-task` “task unchanged” write nothing. Normal stdout contracts
  unchanged (path / message on stdout; follow-ups never on stdout).
- **Shell process cwd** — at process start the binary captures shell process
  cwd (`os.Getwd` / `origWd`, inherited from the interactive shell — not
  merely the optional `<dir>` workDir / source path). Gates compare this
  captured path, so `cd ~ && wrk ~/code/myrepo` uses home as the gate input
  even though create resolves work from the repo path argument.
- **Create follow-up (home gate)** — after successful create, write
  `cd <new-worktree>` **only if** shell process cwd equals the user home
  directory from `os.UserHomeDir()` (OS home API; not a direct
  `os.Getenv("HOME")` in our code). Exact home only — a subdirectory under
  home does **not** qualify. Path compare uses Clean + EvalSymlinks when
  possible. If home cannot be resolved → do **not** write (fail closed).
  Create from a main repo (or any non-home cwd) still succeeds and prints
  the worktree path on stdout; the follow-up file stays empty so the shell
  is not yanked. Tests set `HOME=FakeHome` so `os.UserHomeDir()` resolves
  to FakeHome — FakeHome is the correct “user home” in leaves.
- **Shell-cwd gate (done / set-task move)** — after a successful `--done`
  remove or path-changing `--set-task` move, write the follow-up **only if
  the captured shell cwd no longer exists** (e.g. `os.Stat` not-exist). If
  the path still exists (user was in a sibling worktree, main checkout, or
  other still-valid directory), the follow-up file stays empty so the
  wrapper does not yank the shell. Dest when writing: main repo
  (`TargetPath`) for `--done`; `newPath` for `--set-task` move. Unchanged
  by the create home-gate feature.
- **`--no-cd`** — real binary flag (listed in help and `--complete` flags);
  when set, never write follow-ups even if `WRK_FOLLOWUP_FILE` is set.
  Harmless no-op for follow-ups on other modes. Independent of the home gate.
- **`WRK_FOLLOWUP_FILE`** — ephemeral path set only by the bash `wrk()`
  wrapper when auto-cd is enabled. Binary may write `cd /abs` lines there.
- **`WRK_AUTO_CD`** — user env; `0` makes the wrapper skip temp file creation,
  skip exporting `WRK_FOLLOWUP_FILE`, and skip executing follow-ups.
- **`wrk()` bash wrapper** — defined in `$WRK_HOME/integration/bash.sh`
  (alongside `complete -F _wrk wrk`). Creates temp follow-up file when
  auto-cd on and `--no-cd` not in args; runs `command wrk "$@"`; prints each
  follow-up line to **stderr** exactly as `cd /path` (no `wrk:` prefix);
  executes only whitelisted `cd` + single absolute path via builtin `cd`;
  on `cd` failure the wrapper returns non-zero even if the binary exited 0;
  otherwise returns the binary’s exit status. Removes the temp file.
  Wrapper stays dumb — home gate and cwd-existence gates are binary-side only.
- **Follow-up line format** — one command per line: `cd /absolute/path`.
  Absolute paths only; wrapper never `eval`s arbitrary content.
- **WRK_HOME / Fake HOME** — isolated per leaf at `{WorkRoot}/.wrk` and
  `{WorkRoot}/home` for install/profile tests. FakeHome is also the shell
  cwd for create success leaves that expect a home-gated follow-up write
  (`RepoDir`/`StartDir` = FakeHome + positional main-repo arg).
- **WRK_DATE** — fixed to `2026-06-30` for deterministic worktree naming.

## Tree Overview

```
followup-cd/
├── script-surface/                 # print / install / completion surface
│   ├── print-script/
│   │   └── has-wrapper/            # script defines wrk() + follow-up env
│   ├── install/
│   │   ├── writes-wrapper/         # fresh install writes bash.sh wrapper
│   │   └── upgrades-old-script/    # overwrite pre-seeded completion-only script
│   └── complete/
│       └── no-cd-flag/             # --complete lists --no-cd
├── binary-followup/                # WRK_FOLLOWUP_FILE protocol (no shell cd)
│   ├── create/
│   │   ├── success-env-set/        # shell cwd=FakeHome + env → cd <worktree>
│   │   ├── success-cwd-not-home/   # shell cwd=main repo + env → empty
│   │   ├── success-env-unset/      # no side effects without env
│   │   ├── no-cd-flag/             # --no-cd suppresses write (from home)
│   │   └── failure/                # non-git create fails; file empty
│   ├── done/
│   │   ├── remove-success/         # cwd inside operated wt → cd <main>
│   │   ├── sibling-cwd-empty/      # cwd = sibling A; --done B → empty
│   │   └── main-cwd-empty/         # cwd = main; --done B → empty
│   ├── merge-back/
│   │   └── no-followup/            # success but no cd line
│   └── set-task/
│       ├── move-success/           # cwd inside renamed wt → cd <newPath>
│       ├── task-unchanged/         # slug same → empty file
│       └── sibling-cwd-empty/      # cwd = sibling A; rename B → empty
└── wrapper/                        # source bash.sh; end-to-end auto-cd
    ├── create/
    │   ├── auto-cd/                # StartDir=FakeHome → worktree; stderr cd
    │   ├── auto-cd-from-repo-stays/# StartDir=main repo → stay; no stderr cd
    │   ├── auto-cd-off/            # WRK_AUTO_CD=0; cwd unchanged
    │   └── no-cd-flag/             # wrk --no-cd; cwd unchanged
    ├── done/
    │   ├── auto-cd/                # cwd inside operated wt → main
    │   └── sibling-cwd-stays/      # cwd sibling; --done other → stay
    ├── set-task/
    │   ├── auto-cd/                # cwd inside renamed wt → new path
    │   └── sibling-cwd-stays/      # cwd sibling; rename other → stay
    └── bad-path/
        └── cd-fails/               # binary 0 + bad cd → wrapper non-zero
```

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | script-surface/print-script/has-wrapper | `--bash-integration` stdout defines `wrk()`, mentions follow-up env, keeps completion |
| 2 | script-surface/install/writes-wrapper | `--install` writes `bash.sh` with wrapper + follow-up logic |
| 3 | script-surface/install/upgrades-old-script | install overwrites completion-only pre-seed with wrapper script |
| 4 | script-surface/complete/no-cd-flag | `--complete` flag candidates include `--no-cd` |
| 5 | binary-followup/create/success-env-set | shell cwd=FakeHome + `wrk <mainRepo>` + env → follow-up `cd <abs-worktree>` |
| 6 | binary-followup/create/success-cwd-not-home | shell cwd=main repo + env → worktree created; follow-up empty |
| 7 | binary-followup/create/success-env-unset | create without env → follow-up path untouched |
| 8 | binary-followup/create/no-cd-flag | create from home + `--no-cd` + env → empty follow-up |
| 9 | binary-followup/create/failure | non-git create + env → non-zero, empty follow-up |
| 10 | binary-followup/done/remove-success | `--done` from inside operated wt → `cd <main-repo-abs>` |
| 11 | binary-followup/done/sibling-cwd-empty | shell cwd = sibling A; `--done <B>` → empty follow-up |
| 12 | binary-followup/done/main-cwd-empty | shell cwd = main; `--done <B>` → empty follow-up |
| 13 | binary-followup/merge-back/no-followup | `--merge-back` success + env → empty follow-up |
| 14 | binary-followup/set-task/move-success | `--set-task` move from inside wt → `cd <newPath-abs>` |
| 15 | binary-followup/set-task/task-unchanged | same slug + env → empty follow-up |
| 16 | binary-followup/set-task/sibling-cwd-empty | shell cwd = sibling A; rename B → empty follow-up |
| 17 | wrapper/create/auto-cd | StartDir=FakeHome + `wrk <mainRepo>` → stderr `cd`, `pwd` = worktree |
| 18 | wrapper/create/auto-cd-from-repo-stays | StartDir=main repo → no stderr `cd`; FinalPWD stays main |
| 19 | wrapper/create/auto-cd-off | `WRK_AUTO_CD=0` from home → no stderr `cd`, cwd stays home |
| 20 | wrapper/create/no-cd-flag | `wrk --no-cd` from home → no stderr `cd`, cwd stays home |
| 21 | wrapper/done/auto-cd | wrapper `--done` from linked wt → cwd = main repo |
| 22 | wrapper/done/sibling-cwd-stays | wrapper `--done <other>` from sibling → pwd stays; no stderr `cd` |
| 23 | wrapper/set-task/auto-cd | wrapper `--set-task` from inside wt → cwd = new path |
| 24 | wrapper/set-task/sibling-cwd-stays | wrapper rename other from sibling → pwd stays; no stderr `cd` |
| 25 | wrapper/bad-path/cd-fails | follow-up `cd` to missing path → wrapper non-zero, binary 0 |

## How to Run

```sh
doctest vet ./go-pkgs/cmd/wrk/tests/followup-cd
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd

# By surface
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/script-surface
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/wrapper

# Create home-gate surfaces (new non-home leaves RED until implementer lands gate;
# home success leaves may be RED until create uses home-gated write)
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup/create
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/wrapper/create
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup/create/success-cwd-not-home
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/wrapper/create/auto-cd-from-repo-stays

# Done/set-task existence-gate leaves (unchanged by create home gate)
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup/done/sibling-cwd-empty
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup/done/main-cwd-empty
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup/set-task/sibling-cwd-empty
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/wrapper/done/sibling-cwd-stays
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/wrapper/set-task/sibling-cwd-stays

# Single leaf
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/binary-followup/create/success-env-set
doctest test ./go-pkgs/cmd/wrk/tests/followup-cd/wrapper/create/auto-cd
```

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

type Request struct {
	WorkRoot string
	WrkHome  string
	FakeHome string
	RepoDir  string
	MainRepo string
	WtDir    string
	WtBranch string

	// Mode: print | install | complete | binary | wrapper
	Mode string

	CLIArgs []string

	// Follow-up protocol
	FollowupFile   string // path for WRK_FOLLOWUP_FILE content checks
	UseFollowupEnv bool   // export WRK_FOLLOWUP_FILE to the binary
	AutoCD         string // when non-empty, export WRK_AUTO_CD=<value>

	SetTaskDesc string
	SetTaskEnv  string

	CompleteWords []string
	CompleteCWord int

	// Wrapper
	StartDir       string // shell cwd for wrapper; default RepoDir
	UseFakeWrk     bool   // put a stub wrk on PATH ahead of real binary
	FakeFollowupCD string // absolute path the stub writes as cd target
	PreExistingBashSh string
}

type Response struct {
	ExitCode int
	Stdout   string
	Stderr   string

	Home          string
	WrkHome       string
	BashShPath    string
	BashShContent string

	FollowupPath    string
	FollowupContent string
	FollowupExists  bool

	FinalPWD string
}

func Run(t *testing.T, req *Request) (*Response, error) {
	if req.WorkRoot == "" || req.WrkHome == "" || req.FakeHome == "" {
		return nil, fmt.Errorf("Setup must initialize WorkRoot, WrkHome, and FakeHome")
	}

	if err := seedPreExistingState(req); err != nil {
		return nil, err
	}

	bin := getWrkBin(t)
	resp := &Response{
		Home:       req.FakeHome,
		WrkHome:    req.WrkHome,
		BashShPath: bashShPath(req.WrkHome),
	}

	switch req.Mode {
	case "print":
		stdout, stderr, code, err := runWrkOnce(t, req, bin, []string{"--bash-integration"}, "")
		if err != nil {
			return nil, err
		}
		resp.Stdout, resp.Stderr, resp.ExitCode = stdout, stderr, code
	case "install":
		stdout, stderr, code, err := runWrkOnce(t, req, bin, []string{"--bash-integration", "--install"}, "")
		if err != nil {
			return nil, err
		}
		resp.Stdout, resp.Stderr, resp.ExitCode = stdout, stderr, code
		resp.BashShContent, _ = readFileIfExists(resp.BashShPath)
	case "complete":
		args := []string{"--bash-integration", "--complete", "--"}
		args = append(args, req.CompleteWords...)
		args = append(args, fmt.Sprintf("%d", req.CompleteCWord))
		stdout, stderr, code, err := runWrkOnce(t, req, bin, args, "")
		if err != nil {
			return nil, err
		}
		resp.Stdout, resp.Stderr, resp.ExitCode = stdout, stderr, code
	case "binary":
		if err := prepareFollowupFile(req); err != nil {
			return nil, err
		}
		args := buildBinaryArgs(req)
		followEnv := ""
		if req.UseFollowupEnv {
			if req.FollowupFile == "" {
				return nil, fmt.Errorf("UseFollowupEnv requires FollowupFile")
			}
			followEnv = req.FollowupFile
		}
		stdout, stderr, code, err := runWrkOnce(t, req, bin, args, followEnv)
		if err != nil {
			return nil, err
		}
		resp.Stdout, resp.Stderr, resp.ExitCode = stdout, stderr, code
		captureFollowup(resp, req.FollowupFile)
	case "wrapper":
		if err := ensureIntegrationScript(t, req, bin); err != nil {
			return nil, err
		}
		resp.BashShContent, _ = readFileIfExists(bashShPath(req.WrkHome))
		return runWrapper(t, req, bin, resp)
	default:
		return nil, fmt.Errorf("unknown Mode %q", req.Mode)
	}
	return resp, nil
}

func buildBinaryArgs(req *Request) []string {
	if len(req.CLIArgs) > 0 {
		return append([]string(nil), req.CLIArgs...)
	}
	if req.SetTaskDesc != "" {
		return []string{"--set-task", req.SetTaskDesc}
	}
	return nil
}

func prepareFollowupFile(req *Request) error {
	if req.FollowupFile == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(req.FollowupFile), 0o755); err != nil {
		return err
	}
	// Always start empty so assertions can detect writes.
	return os.WriteFile(req.FollowupFile, nil, 0o644)
}

func captureFollowup(resp *Response, path string) {
	if path == "" {
		return
	}
	resp.FollowupPath = path
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			resp.FollowupExists = false
			return
		}
		return
	}
	resp.FollowupExists = true
	resp.FollowupContent = string(data)
}

func ensureIntegrationScript(t *testing.T, req *Request, bin string) error {
	t.Helper()
	// Prefer install so upgrades rewrite bash.sh; falls back if install fails
	// by writing printed script (still exercises whatever current binary emits).
	_, _, code, err := runWrkOnce(t, req, bin, []string{"--bash-integration", "--install"}, "")
	if err != nil {
		return err
	}
	if code != 0 {
		stdout, _, pcode, perr := runWrkOnce(t, req, bin, []string{"--bash-integration"}, "")
		if perr != nil {
			return perr
		}
		if pcode != 0 {
			return fmt.Errorf("print bash-integration exit %d", pcode)
		}
		if err := os.MkdirAll(filepath.Dir(bashShPath(req.WrkHome)), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(bashShPath(req.WrkHome), []byte(stdout), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func runWrapper(t *testing.T, req *Request, bin string, resp *Response) (*Response, error) {
	t.Helper()
	startDir := req.StartDir
	if startDir == "" {
		startDir = req.RepoDir
	}
	if startDir == "" {
		startDir = req.WorkRoot
	}
	pwdFile := filepath.Join(req.WorkRoot, "final-pwd.txt")
	binDir := filepath.Join(req.WorkRoot, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		return nil, err
	}

	pathPrefix := binDir
	if req.UseFakeWrk {
		fake := filepath.Join(binDir, "wrk")
		target := req.FakeFollowupCD
		if target == "" {
			target = filepath.Join(req.WorkRoot, "no-such-followup-target")
		}
		body := fmt.Sprintf("#!/usr/bin/env bash\n"+
			"if [[ -n \"${WRK_FOLLOWUP_FILE:-}\" ]]; then\n"+
			"  printf 'cd %%s\\n' %q > \"$WRK_FOLLOWUP_FILE\"\n"+
			"fi\n"+
			"exit 0\n", target)
		if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
			return nil, err
		}
	} else {
		// Expose real binary as "wrk" on PATH so `command wrk` resolves.
		link := filepath.Join(binDir, "wrk")
		_ = os.Remove(link)
		if err := os.Symlink(bin, link); err != nil {
			// Fall back to PATH including bin's directory.
			pathPrefix = filepath.Dir(bin) + string(os.PathListSeparator) + binDir
			link = filepath.Join(filepath.Dir(bin), "wrk")
			if filepath.Base(bin) != "wrk" {
				return nil, fmt.Errorf("expected wrk binary named wrk, got %s", bin)
			}
			_ = link
		}
	}

	bashSh := bashShPath(req.WrkHome)
	script := fmt.Sprintf(`
set -euo pipefail
source %q
export HOME=%q
export WRK_HOME=%q
export WRK_DATE=%q
export PATH=%q:"$PATH"
cd %q
set +e
wrk %s
status=$?
set -e
pwd > %q
exit "$status"
`,
		bashSh,
		req.FakeHome,
		req.WrkHome,
		wrkDate,
		pathPrefix,
		startDir,
		shellJoinArgs(buildWrapperArgs(req)),
		pwdFile,
	)

	cmd := exec.Command("bash", "-c", script)
	cmd.Dir = startDir
	env := []string{
		"HOME=" + req.FakeHome,
		"WRK_HOME=" + req.WrkHome,
		"WRK_DATE=" + wrkDate,
		"PATH=" + pathPrefix + string(os.PathListSeparator) + os.Getenv("PATH"),
		"TMPDIR=" + req.WorkRoot,
	}
	if req.AutoCD != "" {
		env = append(env, "WRK_AUTO_CD="+req.AutoCD)
	}
	if req.SetTaskEnv != "" {
		env = append(env, req.SetTaskEnv)
	}
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	exitCode := 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return nil, runErr
		}
	}
	resp.Stdout = outBuf.String()
	resp.Stderr = errBuf.String()
	resp.ExitCode = exitCode
	if data, err := os.ReadFile(pwdFile); err == nil {
		resp.FinalPWD = strings.TrimSpace(string(data))
	}
	return resp, nil
}

func buildWrapperArgs(req *Request) []string {
	if len(req.CLIArgs) > 0 {
		return append([]string(nil), req.CLIArgs...)
	}
	if req.SetTaskDesc != "" {
		return []string{"--set-task", req.SetTaskDesc}
	}
	return nil
}

func shellJoinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	parts := make([]string, len(args))
	for i, a := range args {
		parts[i] = shellQuote(a)
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'"'"'`) + "'"
}

func runWrkOnce(t *testing.T, req *Request, bin string, args []string, followupFile string) (stdout, stderr string, exitCode int, err error) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	if cmd.Dir == "" {
		cmd.Dir = req.WorkRoot
	}
	env := []string{
		"HOME=" + req.FakeHome,
		"WRK_HOME=" + req.WrkHome,
		"WRK_DATE=" + wrkDate,
		"PATH=" + os.Getenv("PATH"),
	}
	if followupFile != "" {
		env = append(env, "WRK_FOLLOWUP_FILE="+followupFile)
	}
	if req.SetTaskEnv != "" {
		env = append(env, req.SetTaskEnv)
	}
	if req.AutoCD != "" {
		env = append(env, "WRK_AUTO_CD="+req.AutoCD)
	}
	cmd.Env = env
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	runErr := cmd.Run()
	exitCode = 0
	if runErr != nil {
		if ee, ok := runErr.(*exec.ExitError); ok {
			exitCode = ee.ExitCode()
		} else {
			return "", "", -1, runErr
		}
	}
	return outBuf.String(), errBuf.String(), exitCode, nil
}

func getWrkBin(t *testing.T) string {
	t.Helper()
	base := os.Getenv("DOCTEST_FIXTURE_ROOT")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		base = filepath.Join(home, "Library", "Caches", "doctest", "fixtures")
	}
	sessionRoot := filepath.Join(base, DOCTEST_SESSION_ID)
	bin := filepath.Join(sessionRoot, "bin", "wrk")
	if _, err := os.Stat(bin); err == nil {
		return bin
	}
	lockPath := filepath.Join(sessionRoot, "bin", ".lock")
	withFlock(t, lockPath, func() {
		if _, err := os.Stat(bin); err == nil {
			return
		}
		// DOCTEST_ROOT = .../cmd/wrk/tests/followup-cd → module main at .../cmd/wrk
		modRoot := filepath.Join(DOCTEST_ROOT, "..", "..")
		cmd := exec.Command("go", "build", "-o", bin, ".")
		cmd.Dir = modRoot
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("build wrk: %v\n%s", err, out)
		}
	})
	return bin
}

func withFlock(t *testing.T, lockPath string, fn func()) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		t.Fatalf("mkdir lock dir: %v", err)
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		t.Fatalf("open lock %s: %v", lockPath, err)
	}
	defer f.Close()
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatalf("flock %s: %v", lockPath, err)
	}
	defer func() { _ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) }()
	fn()
}

func seedPreExistingState(req *Request) error {
	if req.PreExistingBashSh == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(bashShPath(req.WrkHome)), 0o755); err != nil {
		return err
	}
	return os.WriteFile(bashShPath(req.WrkHome), []byte(req.PreExistingBashSh), 0o644)
}
```
