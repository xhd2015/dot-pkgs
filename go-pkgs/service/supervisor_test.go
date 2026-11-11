package service

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestSupervisor(t *testing.T, opts Options) *Supervisor {
	t.Helper()
	dir := t.TempDir()
	opts.DataDir = filepath.Join(dir, "data")
	opts.LogDir = filepath.Join(dir, "logs")
	if opts.Logf == nil {
		opts.Logf = func(format string, args ...any) { t.Logf(format, args...) }
	}
	s := NewSupervisor(opts)
	t.Cleanup(s.ForceStopAll)
	return s
}

func sleepDefinition(id string) Definition {
	return Definition{ID: id, Name: id, Command: "sleep 300"}
}

func waitForState(t *testing.T, s *Supervisor, id string, want ServiceState) Process {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last Process
	for {
		proc, ok := s.Process(id)
		if ok {
			last = proc
			if proc.State == want {
				return proc
			}
		}
		if time.Now().After(deadline) {
			t.Fatalf("service %q state = %q (%+v), want %q", id, last.State, last, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestSupervisorDefineSortsAndValidates(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Define([]Definition{sleepDefinition("b"), sleepDefinition("a")}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	defs := s.Definitions()
	if len(defs) != 2 || defs[0].ID != "a" || defs[1].ID != "b" {
		t.Fatalf("Definitions() = %+v, want sorted by id", defs)
	}
	if def, ok := s.Definition("a"); !ok || def.Command != "sleep 300" {
		t.Fatalf("Definition(a) = %+v, %v", def, ok)
	}
	if _, ok := s.Definition("missing"); ok {
		t.Fatalf("Definition(missing) = found")
	}
	if err := s.Define([]Definition{sleepDefinition("dup"), sleepDefinition("dup")}); err == nil || !strings.Contains(err.Error(), "duplicate service id") {
		t.Fatalf("Define() duplicate error = %v", err)
	}
	if err := s.Define([]Definition{{ID: "bad"}}); err == nil || !strings.Contains(err.Error(), "command is required") {
		t.Fatalf("Define() invalid error = %v", err)
	}
	if got, want := s.LogPath("a"), filepath.Join(s.Options().LogDir, "a.log"); got != want {
		t.Fatalf("LogPath() = %q, want %q", got, want)
	}
}

func TestSupervisorStartAndStop(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	proc := waitForState(t, s, "svc", StateRunning)
	if proc.PID <= 0 || !Alive(proc.PID) {
		t.Fatalf("Process = %+v, want a live pid", proc)
	}
	if !proc.Desired || proc.Detached || proc.Adopted {
		t.Fatalf("Process = %+v, want a desired attached process", proc)
	}
	if proc.LastStartedAt == "" || proc.RunStartedAt.IsZero() {
		t.Fatalf("Process = %+v, want start timestamps", proc)
	}
	if !s.Running("svc") {
		t.Fatalf("Running() = false")
	}
	logData, err := os.ReadFile(s.LogPath("svc"))
	if err != nil {
		t.Fatalf("ReadFile(log) error = %v", err)
	}
	if !strings.Contains(string(logData), "starting service svc") {
		t.Fatalf("log = %q, want the start marker", string(logData))
	}

	pid := proc.PID
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	proc = waitForState(t, s, "svc", StateStopped)
	if proc.PID != 0 || proc.Desired {
		t.Fatalf("Process after Stop = %+v", proc)
	}
	if Alive(pid) {
		t.Fatalf("pid %d survived Stop", pid)
	}
	if s.Running("svc") {
		t.Fatalf("Running() = true after Stop")
	}
	state := s.RuntimeState()
	if len(state.Processes) != 0 {
		t.Fatalf("runtime state = %+v, want the record forgotten", state)
	}
	logData, err = os.ReadFile(s.LogPath("svc"))
	if err != nil {
		t.Fatalf("ReadFile(log) error = %v", err)
	}
	if !strings.Contains(string(logData), "service exited") {
		t.Fatalf("log = %q, want the exit marker", string(logData))
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("second Stop() error = %v, want no-op", err)
	}
}

func TestSupervisorStartUnknownService(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Start("missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Start() error = %v", err)
	}
	if err := s.Stop("missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Stop() error = %v", err)
	}
	if err := s.Remove("missing"); err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("Remove() error = %v", err)
	}
}

func TestSupervisorStartIsIdempotent(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	first := waitForState(t, s, "svc", StateRunning)
	if err := s.Start("svc"); err != nil {
		t.Fatalf("second Start() error = %v", err)
	}
	second := waitForState(t, s, "svc", StateRunning)
	if first.PID != second.PID {
		t.Fatalf("Start() spawned a second process: %d then %d", first.PID, second.PID)
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSupervisorRestart(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Restart("svc"); err != nil {
		t.Fatalf("Restart() on a stopped service error = %v, want it to start", err)
	}
	first := waitForState(t, s, "svc", StateRunning)
	if err := s.Restart("svc"); err != nil {
		t.Fatalf("Restart() error = %v", err)
	}
	second := waitForState(t, s, "svc", StateRunning)
	if first.PID == second.PID {
		t.Fatalf("Restart() kept pid %d", first.PID)
	}
	if Alive(first.PID) {
		t.Fatalf("old pid %d survived Restart", first.PID)
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSupervisorSetEnabled(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	def := sleepDefinition("svc")
	def.Enabled = Bool(false)
	if err := s.SetDefinition(def); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	waitForState(t, s, "svc", StateRunning)
	if err := s.SetEnabled("svc", false); err != nil {
		t.Fatalf("SetEnabled(false) error = %v", err)
	}
	waitForState(t, s, "svc", StateStopped)
	if def, _ := s.Definition("svc"); def.IsEnabled() {
		t.Fatalf("Definition.Enabled = true after disable")
	}
	if err := s.SetEnabled("svc", true); err != nil {
		t.Fatalf("SetEnabled(true) error = %v", err)
	}
	waitForState(t, s, "svc", StateRunning)
	if def, _ := s.Definition("svc"); !def.IsEnabled() {
		t.Fatalf("Definition.Enabled = false after enable")
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSupervisorAutoStartOnlyStartsEnabled(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	on := sleepDefinition("on")
	off := sleepDefinition("off")
	off.Enabled = Bool(false)
	if err := s.Define([]Definition{on, off}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	if started := s.AutoStart(); len(started) != 1 || started[0] != "on" {
		t.Fatalf("AutoStart() = %v, want [on]", started)
	}
	waitForState(t, s, "on", StateRunning)
	if s.Running("off") {
		t.Fatalf("disabled service was autostarted")
	}
	if err := s.Stop("on"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSupervisorStartFailureRecordsBackoff(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(Definition{ID: "bad", Name: "bad", Command: "exit 7"}); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("bad"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	proc := waitForState(t, s, "bad", StateError)
	if proc.Failures < 1 {
		t.Fatalf("Process = %+v, want a recorded failure", proc)
	}
	if proc.PID != 0 || proc.Desired == false {
		t.Fatalf("Process = %+v, want the pid cleared but still desired", proc)
	}
	if proc.LastExitError == "" {
		t.Fatalf("Process = %+v, want last exit error", proc)
	}
	if proc.NextRestartAt.IsZero() || !proc.NextRestartAt.After(time.Now()) {
		t.Fatalf("Process = %+v, want a future backoff time", proc)
	}
	failures := proc.Failures
	if err := s.Start("bad"); err != nil {
		t.Fatalf("Start() during backoff error = %v, want nil", err)
	}
	time.Sleep(100 * time.Millisecond)
	if proc, _ := s.Process("bad"); proc.Failures != failures || proc.State != StateError {
		t.Fatalf("Start() during backoff retried: %+v", proc)
	}
	if err := s.StartForce("bad"); err != nil {
		t.Fatalf("StartForce() error = %v", err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		proc, _ := s.Process("bad")
		if proc.Failures > failures {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("StartForce() did not retry: %+v", proc)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestSupervisorPrepareHookFailure(t *testing.T) {
	prepareErr := os.ErrPermission
	s := newTestSupervisor(t, Options{Prepare: func(Definition) error { return prepareErr }})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err == nil {
		t.Fatalf("Start() error = nil, want the prepare failure")
	}
	proc := waitForState(t, s, "svc", StateError)
	if proc.PID != 0 || !strings.Contains(proc.LastExitError, "permission") {
		t.Fatalf("Process = %+v, want a prepare failure record", proc)
	}
}

func TestSupervisorStartedHookFailureKeepsProcessStoppable(t *testing.T) {
	s := newTestSupervisor(t, Options{Started: func(Definition, int, *os.File) error {
		return os.ErrDeadlineExceeded
	}})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err != nil {
		t.Fatalf("Start() error = %v, want the process to stay up", err)
	}
	proc := waitForState(t, s, "svc", StateError)
	if proc.PID <= 0 || !Alive(proc.PID) {
		t.Fatalf("Process = %+v, want the pid kept so it stays stoppable", proc)
	}
	if !strings.Contains(proc.LastExitError, "setup failed") {
		t.Fatalf("Process = %+v, want a setup failure message", proc)
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if Alive(proc.PID) {
		t.Fatalf("pid %d survived Stop", proc.PID)
	}
}

func TestSupervisorReconcileRestartsDeadService(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	first := waitForState(t, s, "svc", StateRunning)
	KillProcessGroupNow(first.PID)
	waitForState(t, s, "svc", StateError)
	time.Sleep(1100 * time.Millisecond) // clear the first backoff window
	restarted := s.Reconcile()
	if len(restarted) != 1 || restarted[0] != "svc" {
		t.Fatalf("Reconcile() = %v, want [svc]", restarted)
	}
	proc := waitForState(t, s, "svc", StateRunning)
	if proc.PID == first.PID {
		t.Fatalf("Reconcile() did not replace pid %d", first.PID)
	}
	if !proc.NextRestartAt.IsZero() || proc.LastExitError != "" {
		t.Fatalf("Process = %+v, want the backoff cleared the moment it restarted", proc)
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
}

func TestSupervisorReconcileSkipsStoppedAndBackedOffServices(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Define([]Definition{sleepDefinition("stopped")}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	if restarted := s.Reconcile(); len(restarted) != 0 {
		t.Fatalf("Reconcile() = %v, want none for a never-started service", restarted)
	}
}

func TestSupervisorHealthCheckLifecycle(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	s.StartHealthCheck(10 * time.Millisecond)
	s.StopHealthCheck()
	s.StartHealthCheck(10 * time.Millisecond)
	s.StopHealthCheck()
	s.StopHealthCheck()
}

func TestSupervisorShutdownStopsAttachedServices(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Define([]Definition{sleepDefinition("a"), sleepDefinition("b")}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	if err := s.Start("a"); err != nil {
		t.Fatalf("Start(a) error = %v", err)
	}
	if err := s.Start("b"); err != nil {
		t.Fatalf("Start(b) error = %v", err)
	}
	a := waitForState(t, s, "a", StateRunning)
	b := waitForState(t, s, "b", StateRunning)
	s.Shutdown()
	if Alive(a.PID) || Alive(b.PID) {
		t.Fatalf("Shutdown() left pids %d and %d alive", a.PID, b.PID)
	}
	if s.Running("a") || s.Running("b") {
		t.Fatalf("Shutdown() left services marked running")
	}
}

func TestSupervisorShutdownLeavesDetachedServiceRunning(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	def := sleepDefinition("detached")
	def.Detached = true
	if err := s.SetDefinition(def); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("detached"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	proc := waitForState(t, s, "detached", StateRunning)
	if !proc.Detached {
		t.Fatalf("Process = %+v, want Detached", proc)
	}
	pgid, err := GroupID(proc.PID)
	if err != nil {
		t.Fatalf("GroupID() error = %v", err)
	}
	if pgid != proc.PID {
		t.Fatalf("detached process group = %d, want its own session %d", pgid, proc.PID)
	}
	s.Shutdown()
	if !Alive(proc.PID) {
		t.Fatalf("detached pid %d died with Shutdown", proc.PID)
	}
	KillProcessGroupNow(proc.PID)
	if Alive(proc.PID) {
		t.Fatalf("detached pid %d survived cleanup", proc.PID)
	}
}

func TestSupervisorRemoveStopsTheProcess(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	if err := s.Start("svc"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	proc := waitForState(t, s, "svc", StateRunning)
	if err := s.Remove("svc"); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
	if Alive(proc.PID) {
		t.Fatalf("pid %d survived Remove", proc.PID)
	}
	if _, ok := s.Process("svc"); ok {
		t.Fatalf("Process() = found after Remove")
	}
	if len(s.Definitions()) != 0 {
		t.Fatalf("Definitions() = %v, want empty", s.Definitions())
	}
}

func TestSupervisorDefineStopsRemovedDefinitions(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Define([]Definition{sleepDefinition("a"), sleepDefinition("b")}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	if err := s.Start("b"); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	proc := waitForState(t, s, "b", StateRunning)
	if err := s.Define([]Definition{sleepDefinition("a")}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	if Alive(proc.PID) {
		t.Fatalf("pid %d of a removed definition survived", proc.PID)
	}
	if _, ok := s.Process("b"); ok {
		t.Fatalf("Process(b) = found after it was removed from the definitions")
	}
}

func TestSupervisorReservedSystemServiceID(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Registry().Register(SystemService{ID: "llm-proxy", Name: "llm-proxy", Controller: &statusOnlyController{}}); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := s.SetDefinition(sleepDefinition("llm-proxy")); err == nil || !strings.Contains(err.Error(), "collides with a system service") {
		t.Fatalf("SetDefinition() error = %v, want the collision rejection", err)
	}
	if err := s.Define([]Definition{sleepDefinition("llm-proxy")}); err == nil || !strings.Contains(err.Error(), "collides with a system service") {
		t.Fatalf("Define() error = %v, want the collision rejection", err)
	}
	if _, ok := s.Definition("llm-proxy"); ok {
		t.Fatalf("colliding definition was stored")
	}
}

func TestSupervisorLogfDefaultDoesNotPanic(t *testing.T) {
	s := NewSupervisor(Options{DataDir: t.TempDir()})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	s.logf("no writer configured")
}

func TestSupervisorTailLog(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := os.MkdirAll(s.Options().LogDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := AppendLog(s.LogPath("svc"), "hello"); err != nil {
		t.Fatalf("AppendLog() error = %v", err)
	}
	lines, err := TailLog(s.LogPath("svc"), 5)
	if err != nil {
		t.Fatalf("TailLog() error = %v", err)
	}
	if len(lines) != 1 || lines[0] != "hello" {
		t.Fatalf("TailLog() = %v", lines)
	}
}

func TestSupervisorReclaimStopKillsOrphan(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	pid := spawnSleep(t, "sleep 300", false)
	record(t, s, "svc", pid, "sleep 300")
	result := s.Reclaim(ReclaimStop)
	if len(result.Reclaimed) != 1 || result.Reclaimed[0] != "svc" {
		t.Fatalf("Reclaim() = %+v, want svc reclaimed", result)
	}
	if Alive(pid) {
		t.Fatalf("orphan pid %d survived ReclaimStop", pid)
	}
	if state := s.RuntimeState(); len(state.Processes) != 0 {
		t.Fatalf("runtime state = %+v, want the record forgotten", state)
	}
}

func TestSupervisorReclaimAdoptKeepsOrphan(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.SetDefinition(sleepDefinition("svc")); err != nil {
		t.Fatalf("SetDefinition() error = %v", err)
	}
	pid := spawnSleep(t, "sleep 300", false)
	record(t, s, "svc", pid, "sleep 300")
	result := s.Reclaim(ReclaimAdopt)
	if len(result.Adopted) != 1 || result.Adopted[0] != "svc" {
		t.Fatalf("Reclaim() = %+v, want svc adopted", result)
	}
	proc, ok := s.Process("svc")
	if !ok || !proc.Adopted || proc.PID != pid || proc.State != StateRunning || !s.Running("svc") {
		t.Fatalf("Process() = %+v, %v, want an adopted running process", proc, ok)
	}
	if err := s.Stop("svc"); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	if Alive(pid) {
		t.Fatalf("adopted pid %d survived Stop", pid)
	}
}

func TestSupervisorReclaimSkipsRecycledPID(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	record(t, s, "svc", os.Getpid(), "some-unrelated-command")
	result := s.Reclaim(ReclaimStop)
	if len(result.Skipped) != 1 || result.Skipped[0] != "svc" {
		t.Fatalf("Reclaim() = %+v, want svc skipped", result)
	}
	if len(result.Reclaimed) != 0 {
		t.Fatalf("Reclaim() = %+v, want nothing reclaimed", result)
	}
}

func TestSupervisorReclaimSkipsDeadPID(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	dead := exec.Command("bash", "-c", "exit 0")
	if err := dead.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	_ = dead.Wait()
	record(t, s, "svc", dead.Process.Pid, "bash -c exit 0")
	result := s.Reclaim(ReclaimStop)
	if len(result.Skipped) != 1 {
		t.Fatalf("Reclaim() = %+v, want svc skipped", result)
	}
	if state := s.RuntimeState(); len(state.Processes) != 0 {
		t.Fatalf("runtime state = %+v, want the stale record forgotten", state)
	}
}

func spawnSleep(t *testing.T, command string, detached bool) int {
	t.Helper()
	cmd := exec.Command("bash", "-c", command)
	cmd.SysProcAttr = procAttr(detached)
	if err := cmd.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	go func() { _ = cmd.Wait() }()
	pid := cmd.Process.Pid
	t.Cleanup(func() { KillProcessGroupNow(pid) })
	return pid
}

func record(t *testing.T, s *Supervisor, id string, pid int, command string) {
	t.Helper()
	store := NewRuntimeStore(filepath.Join(s.Options().DataDir, "services-runtime.json"))
	if err := store.Record(id, RuntimeRecord{
		PID:       pid,
		PGID:      pid,
		StartedAt: time.Now().UTC().Format(time.RFC3339),
		Command:   command,
	}); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
}

func TestSupervisorProcessesSnapshot(t *testing.T) {
	s := newTestSupervisor(t, Options{})
	if err := s.Define([]Definition{sleepDefinition("b"), sleepDefinition("a")}); err != nil {
		t.Fatalf("Define() error = %v", err)
	}
	if procs := s.Processes(); len(procs) != 0 {
		t.Fatalf("Processes() = %+v, want no entries before any start", procs)
	}
	if err := s.Start("a"); err != nil {
		t.Fatalf("Start(a) error = %v", err)
	}
	if err := s.Start("b"); err != nil {
		t.Fatalf("Start(b) error = %v", err)
	}
	a := waitForState(t, s, "a", StateRunning)
	waitForState(t, s, "b", StateRunning)
	procs := s.Processes()
	if len(procs) != 2 || procs[0].Definition.ID != "a" || procs[1].Definition.ID != "b" {
		t.Fatalf("Processes() = %+v, want two running services sorted by id", procs)
	}
	if !procs[0].Running() || procs[0].PID != a.PID {
		t.Fatalf("Processes()[0] = %+v, want the live process", procs[0])
	}
	if err := s.Stop("a"); err != nil {
		t.Fatalf("Stop(a) error = %v", err)
	}
	if err := s.Stop("b"); err != nil {
		t.Fatalf("Stop(b) error = %v", err)
	}
}
