# Scenario

**Feature**: optional create-mode interceptor via `$WRK_HOME/config.json`

```
# create path: parse → resolve → auto-record → intercept?
config.json create.interceptor enabled + create mode
  -> expand argv/vars -> exec fake tool on PATH
  -> skip native worktree + outer follow-up cd

# escape / inactive
no config | enabled:false | --no-interceptor | WRK_NO_INTERCEPTOR=1 | non-create
  -> native path; interceptor ignored
```

## Preconditions

- Git is available for create and status fixtures.
- Fake interceptor tools live under `{WorkRoot}/interceptor-bin` and are prepended via `Request.PathPrepend`.
- Config is written to `{WRK_HOME}/config.json` only when a leaf enables it.

## Steps

- Descendants initialize a git repo, optional config, and optional fake `kool` on PATH.
- Create leaves run bare `wrk` (or with `-t` / `--no-interceptor`) from the repo or FakeHome.
- Asserts inspect stdout/exit, `{WRK_HOME}/worktrees/`, interceptor argv log, projects.json, events.jsonl, and follow-up files.

## Context

- Default fake tool name: `kool` (matches the docs recipe; not real kool).
- Fake records length-prefixed argv payloads to `FAKE_INTERCEPTOR_LOG` (supports multiline args).
- Fake prints `intercepted\n` on stdout and exits with `FAKE_INTERCEPTOR_EXIT` (default 0).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	fakeInterceptorName    = "kool"
	fakeInterceptorStdout  = "intercepted"
	envFakeInterceptorLog  = "FAKE_INTERCEPTOR_LOG"
	envFakeInterceptorExit = "FAKE_INTERCEPTOR_EXIT"
)

type interceptorConfigFile struct {
	Version int `json:"version"`
	Create  *struct {
		Interceptor *interceptorSpec `json:"interceptor,omitempty"`
	} `json:"create,omitempty"`
}

type interceptorSpec struct {
	Enabled bool                   `json:"enabled"`
	Argv    []string               `json:"argv"`
	Vars    map[string]interface{} `json:"vars,omitempty"`
}

type interceptorProjectFile struct {
	Version  int                   `json:"version"`
	Projects []interceptorProject  `json:"projects"`
}

type interceptorProject struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type interceptorEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	ensureInterceptorHelpersUsed()
	return nil
}

func worktreesRoot(wrkHome string) string {
	return filepath.Join(wrkHome, "worktrees")
}

func configJSONPath(wrkHome string) string {
	return filepath.Join(wrkHome, "config.json")
}

func projectsJSONPathInterceptor(wrkHome string) string {
	return filepath.Join(wrkHome, "projects.json")
}

func eventsJSONLPathInterceptor(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

// setupMainRepoForInterceptor clones a go.mod seed repo as myrepo and points RepoDir at it.
func setupMainRepoForInterceptor(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	return mainRepo
}

// writeInterceptorConfig writes {WRK_HOME}/config.json for create.interceptor tests.
func writeInterceptorConfig(t *testing.T, wrkHome string, enabled bool, argv []string, vars map[string]interface{}) {
	t.Helper()
	cfg := interceptorConfigFile{Version: 1}
	cfg.Create = &struct {
		Interceptor *interceptorSpec `json:"interceptor,omitempty"`
	}{
		Interceptor: &interceptorSpec{
			Enabled: enabled,
			Argv:    argv,
			Vars:    vars,
		},
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("marshal config: %v", err)
	}
	data = append(data, '\n')
	writeFile(t, configJSONPath(wrkHome), string(data))
}

// writeEnabledSimpleInterceptor writes a minimal enabled config that runs kool with fixed argv
// plus optional ${work_dir} so expansion of builtins is exercised on the happy path.
func writeEnabledSimpleInterceptor(t *testing.T, wrkHome string) {
	t.Helper()
	writeInterceptorConfig(t, wrkHome, true, []string{
		fakeInterceptorName, "space", "create", "--work-dir", "${work_dir}",
	}, nil)
}

// writeRecipeInterceptor writes the docs-style argv/vars recipe using fake kool.
func writeRecipeInterceptor(t *testing.T, wrkHome string) {
	t.Helper()
	writeInterceptorConfig(t, wrkHome, true, []string{
		fakeInterceptorName, "macos", "space", "create",
		"--run", "kool", "iterm2", ".", "-n",
		"--send", "${send}",
	}, map[string]interface{}{
		"intent_prompt": "/intent-route ${task}",
		"send": []interface{}{
			"wrk --no-interceptor ${args_shell_safe}",
			"agent-run run --session-id-from-prompt --open --agent-runner=grok-tty ${intent_prompt|shell_safe}",
		},
	})
}

// installFakeInterceptor places a length-prefixed argv logger named name on PathPrepend.
// exitCode is exported via FAKE_INTERCEPTOR_EXIT (0 when exitCode==0).
func installFakeInterceptor(t *testing.T, req *Request, name string, exitCode int) {
	t.Helper()
	if name == "" {
		name = fakeInterceptorName
	}
	binDir := filepath.Join(req.WorkRoot, "interceptor-bin")
	mkdirAll(t, binDir)
	logPath := filepath.Join(req.WorkRoot, "interceptor-argv.log")
	// Truncate prior log so "not invoked" asserts are reliable.
	writeFile(t, logPath, "")
	req.InterceptorLog = logPath
	req.PathPrepend = binDir

	// Portable ARGC/LEN framing: supports args containing spaces and newlines.
	// Log configured argv form: basename($0) as argv[0] (command name), then "$@".
	// Shell's $#/$@ omit the program name; shebang $0 is often an absolute path.
	body := `#!/bin/sh
log="${FAKE_INTERCEPTOR_LOG:-}"
if [ -n "$log" ]; then
  {
    # Reconstruct argv as [cmd_name, args...] matching config argv expansion.
    cmd_name=$(basename "$0")
    printf 'ARGC %s\n' "$(($# + 1))"
    for a in "$cmd_name" "$@"; do
      len=$(printf '%s' "$a" | wc -c | tr -d ' \t')
      printf 'LEN %s\n' "$len"
      printf '%s' "$a"
      printf '\n'
    done
  } > "$log"
fi
printf 'intercepted\n'
code="${FAKE_INTERCEPTOR_EXIT:-0}"
exit "$code"
`
	fake := filepath.Join(binDir, name)
	if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake interceptor %s: %v", fake, err)
	}

	req.ExtraEnv = append(req.ExtraEnv,
		envFakeInterceptorLog+"="+logPath,
	)
	if exitCode != 0 {
		req.ExtraEnv = append(req.ExtraEnv, fmt.Sprintf("%s=%d", envFakeInterceptorExit, exitCode))
	}
}

// readInterceptorArgs parses the portable ARGC/LEN framing written by the fake tool.
func readInterceptorArgs(t *testing.T, logPath string) []string {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read interceptor log %s: %v", logPath, err)
	}
	if len(data) == 0 {
		return nil
	}
	// Framing:
	// ARGC N\n
	// LEN L\n
	// <L bytes payload>
	// \n
	// ...
	s := string(data)
	if !strings.HasPrefix(s, "ARGC ") {
		t.Fatalf("interceptor log missing ARGC header: %q", s)
	}
	lines := strings.SplitN(s, "\n", 2)
	if len(lines) < 2 {
		t.Fatalf("interceptor log truncated after ARGC: %q", s)
	}
	var argc int
	if _, err := fmt.Sscanf(lines[0], "ARGC %d", &argc); err != nil {
		t.Fatalf("parse ARGC: %v in %q", err, lines[0])
	}
	rest := []byte(lines[1])
	var args []string
	for i := 0; i < argc; i++ {
		// next line: LEN L\n
		nl := -1
		for j, b := range rest {
			if b == '\n' {
				nl = j
				break
			}
		}
		if nl < 0 {
			t.Fatalf("interceptor log: missing LEN for arg %d", i)
		}
		header := string(rest[:nl])
		var n int
		if _, err := fmt.Sscanf(header, "LEN %d", &n); err != nil {
			t.Fatalf("parse LEN for arg %d: %v in %q", i, err, header)
		}
		rest = rest[nl+1:]
		if len(rest) < n {
			t.Fatalf("interceptor log: arg %d short payload want %d got %d", i, n, len(rest))
		}
		args = append(args, string(rest[:n]))
		rest = rest[n:]
		// optional trailing newline written after payload
		if len(rest) > 0 && rest[0] == '\n' {
			rest = rest[1:]
		}
	}
	return args
}

func assertInterceptorNotInvoked(t *testing.T, req *Request) {
	t.Helper()
	if req.InterceptorLog == "" {
		return
	}
	data, err := os.ReadFile(req.InterceptorLog)
	if err != nil {
		if os.IsNotExist(err) {
			return
		}
		t.Fatalf("read interceptor log: %v", err)
	}
	if len(data) != 0 {
		t.Fatalf("interceptor should not have been invoked; log=%q", string(data))
	}
}

func assertInterceptorInvoked(t *testing.T, req *Request) []string {
	t.Helper()
	if req.InterceptorLog == "" {
		t.Fatal("InterceptorLog empty")
	}
	args := readInterceptorArgs(t, req.InterceptorLog)
	if len(args) == 0 {
		t.Fatal("interceptor should have been invoked with at least one arg")
	}
	return args
}

func listWorktreeEntries(t *testing.T, wrkHome string) []string {
	t.Helper()
	root := worktreesRoot(wrkHome)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read worktrees dir: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

func assertNoWorktreesUnderHome(t *testing.T, wrkHome string) {
	t.Helper()
	names := listWorktreeEntries(t, wrkHome)
	if len(names) != 0 {
		t.Fatalf("expected no worktrees under %s, got %v", worktreesRoot(wrkHome), names)
	}
}

func shellSafeQuote(s string) string {
	// POSIX single-quote encoding: empty → ''; ' → '\''
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func readProjectsInterceptor(t *testing.T, wrkHome string) interceptorProjectFile {
	t.Helper()
	data, err := os.ReadFile(projectsJSONPathInterceptor(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return interceptorProjectFile{}
		}
		t.Fatalf("read projects.json: %v", err)
	}
	var pf interceptorProjectFile
	if err := json.Unmarshal(data, &pf); err != nil {
		t.Fatalf("parse projects.json: %v", err)
	}
	return pf
}

func assertProjectAutoRecorded(t *testing.T, wrkHome, mainRepo string) {
	t.Helper()
	pf := readProjectsInterceptor(t, wrkHome)
	want, err := filepath.EvalSymlinks(mainRepo)
	if err != nil {
		want = mainRepo
	}
	wantAbs, err := filepath.Abs(want)
	if err == nil {
		want = wantAbs
	}
	for _, p := range pf.Projects {
		got := p.Path
		if g, err := filepath.EvalSymlinks(got); err == nil {
			got = g
		}
		if g, err := filepath.Abs(got); err == nil {
			got = g
		}
		if got == want {
			if p.Source != "auto" {
				t.Fatalf("project source: want auto, got %q", p.Source)
			}
			if p.AddedAt == "" {
				t.Fatalf("project missing added_at")
			}
			if _, err := time.Parse(time.RFC3339, p.AddedAt); err != nil {
				t.Fatalf("added_at not RFC3339: %q", p.AddedAt)
			}
			return
		}
	}
	t.Fatalf("projects.json missing auto record for %q; got %+v", want, pf.Projects)
}

func readEventsInterceptor(t *testing.T, wrkHome string) []interceptorEvent {
	t.Helper()
	data, err := os.ReadFile(eventsJSONLPathInterceptor(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []interceptorEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev interceptorEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func assertLastEventCommand(t *testing.T, wrkHome, wantCommand string, wantExit int) {
	t.Helper()
	events := readEventsInterceptor(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one events.jsonl entry")
	}
	ev := events[len(events)-1]
	if ev.Command != wantCommand {
		t.Fatalf("event command: want %q, got %q (events=%+v)", wantCommand, ev.Command, events)
	}
	if ev.ExitCode != wantExit {
		t.Fatalf("event exit_code: want %d, got %d", wantExit, ev.ExitCode)
	}
}

func readFollowupInterceptor(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		t.Fatalf("read follow-up %s: %v", path, err)
	}
	return string(data)
}

func assertFollowupHasNoCD(t *testing.T, req *Request) {
	t.Helper()
	if req.FollowupFile == "" {
		t.Fatal("FollowupFile empty")
	}
	got := readFollowupInterceptor(t, req.FollowupFile)
	if strings.Contains(got, "cd ") || strings.HasPrefix(strings.TrimSpace(got), "cd") {
		t.Fatalf("follow-up file should have no outer cd line, got %q", got)
	}
}

func ensureInterceptorHelpersUsed() {
	_ = worktreesRoot
	_ = configJSONPath
	_ = projectsJSONPathInterceptor
	_ = eventsJSONLPathInterceptor
	_ = setupMainRepoForInterceptor
	_ = writeInterceptorConfig
	_ = writeEnabledSimpleInterceptor
	_ = writeRecipeInterceptor
	_ = installFakeInterceptor
	_ = readInterceptorArgs
	_ = assertInterceptorNotInvoked
	_ = assertInterceptorInvoked
	_ = listWorktreeEntries
	_ = assertNoWorktreesUnderHome
	_ = shellSafeQuote
	_ = readProjectsInterceptor
	_ = assertProjectAutoRecorded
	_ = readEventsInterceptor
	_ = assertLastEventCommand
	_ = readFollowupInterceptor
	_ = assertFollowupHasNoCD
	_ = fakeInterceptorStdout
}
```
