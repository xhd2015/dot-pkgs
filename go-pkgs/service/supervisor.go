package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// ServiceState is the lifecycle state of one supervised service.
type ServiceState string

const (
	// StateStarting means start is in progress and no pid is recorded yet.
	StateStarting ServiceState = "starting"
	// StateRunning means a live process was recorded.
	StateRunning ServiceState = "running"
	// StateStopped means the service is intentionally not running.
	StateStopped ServiceState = "stopped"
	// StateError means the last start or exit failed.
	StateError ServiceState = "error"
	// StateUnknown means the host has never observed the service.
	StateUnknown ServiceState = "unknown"
)

// DefaultHealthCheckInterval is how often a supervised service that should be
// running is re-checked when no exit was observed (an adopted process, a
// reused pid, or a manager that started later).
const DefaultHealthCheckInterval = 5 * time.Second

// Process is a point-in-time view of one supervised service, safe to serialize
// or render without holding the supervisor's lock.
type Process struct {
	Definition    Definition
	State         ServiceState
	PID           int
	PGID          int
	Desired       bool
	Detached      bool
	Adopted       bool
	LogPath       string
	LastStartedAt string
	LastExitedAt  string
	LastExitError string
	ShellEnvError string
	EffectivePath string
	Failures      int
	NextRestartAt time.Time
	RunStartedAt  time.Time
}

// Running reports whether the snapshot has a pid.
func (p Process) Running() bool { return p.PID > 0 && p.State == StateRunning }

// Options configures a Supervisor. Every hook is optional.
type Options struct {
	// DataDir is where the runtime pid file lives
	// (<DataDir>/services-runtime.json). Empty disables orphan tracking.
	DataDir string
	// LogDir is where per-service logs go. Empty means <DataDir>/services.
	LogDir string
	// Shell is the command prefix used to run definitions, default
	// ["bash", "-lc"].
	Shell []string
	// Now overrides the clock, for tests.
	Now func() time.Time
	// Logf receives supervisor messages. Defaults to fmt.Printf.
	Logf func(format string, args ...any)

	// Env builds the process environment. Nil uses BuildEnv with the
	// definition's env layers; hosts that enrich PATH themselves build on top.
	Env func(Definition) ([]string, error)
	// Prepare runs after the log file is open and before the process starts.
	// Hosts use it to free a configured port.
	Prepare func(Definition) error
	// Started runs after a successful start, with the log file still open.
	// An error is recorded as a runtime error without stopping the process.
	Started func(Definition, int, *os.File) error
	// Stopping runs before the process is signalled.
	Stopping func(Definition)
	// Exited runs after every exit, before any restart.
	Exited func(Definition)
}

type serviceProcess struct {
	def           Definition
	cmd           *exec.Cmd
	pid           int
	pgid          int
	state         ServiceState
	desired       bool
	adopted       bool
	stopRequested bool
	logPath       string

	lastStartedAt string
	lastExitedAt  string
	lastExitError string
	shellEnvError string
	effectivePath string

	failures      int
	nextRestartAt time.Time
	runStartedAt  time.Time
}

// Supervisor runs child-process services for a host: start, stop, restart,
// crash backoff, logs, and adoption of processes a previous host instance left
// behind.
//
// All methods are safe for concurrent use.
type Supervisor struct {
	opts  Options
	shell []string
	now   func() time.Time

	runtime *RuntimeStore

	mu       sync.Mutex
	defs     map[string]Definition
	order    []string
	entries  map[string]*serviceProcess
	started  bool
	stopCh   chan struct{}
	registry *SystemRegistry
}

// NewSupervisor returns a supervisor for an initially empty definition set.
func NewSupervisor(opts Options) *Supervisor {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.LogDir == "" {
		if opts.DataDir != "" {
			opts.LogDir = filepath.Join(opts.DataDir, "services")
		} else {
			// A host without a data directory still needs a log file outside
			// its cwd, which may not be writable.
			opts.LogDir = filepath.Join(os.TempDir(), "dot-pkgs-service-logs")
		}
	}
	shell := opts.Shell
	if len(shell) == 0 {
		shell = []string{"bash", "-lc"}
	}
	var runtime *RuntimeStore
	if opts.DataDir != "" {
		runtime = NewRuntimeStore(filepath.Join(opts.DataDir, "services-runtime.json"))
	}
	return &Supervisor{
		opts:     opts,
		shell:    shell,
		now:      opts.Now,
		runtime:  runtime,
		defs:     map[string]Definition{},
		entries:  map[string]*serviceProcess{},
		registry: NewSystemRegistry(),
	}
}

// Options returns the supervisor's configuration.
func (s *Supervisor) Options() Options { return s.opts }

// RuntimeState returns the persisted pid records. Hosts use it for diagnostics;
// Reclaim is the supported way to reconcile them.
func (s *Supervisor) RuntimeState() RuntimeState {
	if s.runtime == nil {
		return emptyRuntimeState()
	}
	return s.runtime.Load()
}

// Registry returns the system-service registry that shares this supervisor's
// lifecycle. Hosts register embedded services here so they appear next to
// child-process services.
func (s *Supervisor) Registry() *SystemRegistry { return s.registry }

func (s *Supervisor) logf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if s.opts.Logf != nil {
		s.opts.Logf("%s", message)
		return
	}
	if !strings.HasSuffix(message, "\n") {
		message += "\n"
	}
	fmt.Fprint(os.Stderr, message)
}

// Define replaces the definition set.
//
// Live processes of definitions that remain keep running with their state; a
// definition that disappeared is stopped first, because leaving it running
// would make it invisible and unreclaimable.
func (s *Supervisor) Define(defs []Definition) error {
	normalized := make([]Definition, 0, len(defs))
	seen := make(map[string]struct{}, len(defs))
	for _, def := range defs {
		def, err := NormalizeDefinition(def)
		if err != nil {
			return err
		}
		if err := ValidateDefinition(def); err != nil {
			return err
		}
		if _, ok := seen[def.ID]; ok {
			return fmt.Errorf("duplicate service id: %s", def.ID)
		}
		seen[def.ID] = struct{}{}
		normalized = append(normalized, def)
	}
	SortDefinitions(normalized)
	for _, def := range normalized {
		if err := s.rejectSystemServiceID(def.ID); err != nil {
			return err
		}
	}

	s.mu.Lock()
	defsByID := make(map[string]Definition, len(normalized))
	order := make([]string, 0, len(normalized))
	for _, def := range normalized {
		defsByID[def.ID] = def
		order = append(order, def.ID)
	}
	var removed []string
	for id := range s.entries {
		if _, ok := defsByID[id]; !ok {
			removed = append(removed, id)
		}
	}
	s.mu.Unlock()

	for _, id := range removed {
		s.logf("[service] definition %q was removed; stopping its process", id)
		if err := s.stop(id, true); err != nil {
			s.logf("[service] failed to stop removed service %q: %v", id, err)
		}
		s.mu.Lock()
		delete(s.entries, id)
		s.mu.Unlock()
	}

	s.mu.Lock()
	s.defs = defsByID
	s.order = order
	for id, entry := range s.entries {
		if def, ok := defsByID[id]; ok {
			entry.def = def
		}
	}
	s.mu.Unlock()
	return nil
}

// rejectSystemServiceID refuses a user definition that reuses a system service
// id. The two share one listing in hosts, so a collision would give one id two
// meanings: a child process and an in-process subsystem.
func (s *Supervisor) rejectSystemServiceID(id string) error {
	if _, ok := s.registry.Get(id); ok {
		return fmt.Errorf("service id %q collides with a system service", id)
	}
	return nil
}

// SetDefinition adds or replaces one definition without touching the others.
func (s *Supervisor) SetDefinition(def Definition) error {
	def, err := NormalizeDefinition(def)
	if err != nil {
		return err
	}
	if err := ValidateDefinition(def); err != nil {
		return err
	}
	if err := s.rejectSystemServiceID(def.ID); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.defs[def.ID]; !ok {
		s.order = append(s.order, def.ID)
		sort.Strings(s.order)
	}
	s.defs[def.ID] = def
	if entry := s.entries[def.ID]; entry != nil {
		entry.def = def
	}
	return nil
}

// Remove stops id and forgets its definition.
func (s *Supervisor) Remove(id string) error {
	s.mu.Lock()
	_, hasDef := s.defs[id]
	_, hasEntry := s.entries[id]
	s.mu.Unlock()
	if !hasDef && !hasEntry {
		return fmt.Errorf("service %q not found", id)
	}
	if err := s.stop(id, true); err != nil {
		return err
	}
	s.mu.Lock()
	delete(s.entries, id)
	delete(s.defs, id)
	s.order = removeString(s.order, id)
	s.mu.Unlock()
	return nil
}

// Definitions returns the current definitions in sorted id order.
func (s *Supervisor) Definitions() []Definition {
	s.mu.Lock()
	defer s.mu.Unlock()
	defs := make([]Definition, 0, len(s.order))
	for _, id := range s.order {
		if def, ok := s.defs[id]; ok {
			defs = append(defs, def)
		}
	}
	return defs
}

// Definition returns one definition.
func (s *Supervisor) Definition(id string) (Definition, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	def, ok := s.defs[id]
	return def, ok
}

// LogPath returns the log file of id, whether or not the service ever started.
func (s *Supervisor) LogPath(id string) string {
	s.mu.Lock()
	entry := s.entries[id]
	s.mu.Unlock()
	if entry != nil && entry.logPath != "" {
		return entry.logPath
	}
	if s.opts.LogDir == "" {
		return ""
	}
	return LogPath(s.opts.LogDir, id)
}

// Start starts id unless crash backoff is still pending. A pending backoff
// returns nil with the service in StateError, so a caller that starts
// everything at boot does not report a crash loop as a boot failure.
func (s *Supervisor) Start(id string) error { return s.start(id, false) }

// StartForce starts id, ignoring crash backoff. Used when the user explicitly
// asks to start or restart.
func (s *Supervisor) StartForce(id string) error { return s.start(id, true) }

func (s *Supervisor) start(id string, force bool) error {
	s.mu.Lock()
	def, ok := s.defs[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("service %q not found", id)
	}
	entry := s.entries[id]
	if entry == nil {
		entry = &serviceProcess{def: def, state: StateStopped, logPath: LogPath(s.opts.LogDir, id)}
		s.entries[id] = entry
	} else {
		entry.def = def
	}
	entry.desired = true
	entry.stopRequested = false

	if entry.pid > 0 && Alive(entry.pid) {
		entry.state = StateRunning
		s.mu.Unlock()
		return nil
	}
	if entry.state == StateStarting {
		// Another caller is already spawning this service; a second child would
		// fight for the same port.
		s.mu.Unlock()
		return nil
	}
	if !force && !entry.nextRestartAt.IsZero() && s.now().Before(entry.nextRestartAt) {
		entry.state = StateError
		s.mu.Unlock()
		return nil
	}

	// A pid that is no longer alive belongs to a process this supervisor did
	// not parent (an adopted orphan) or to a crashed child whose exit has not
	// been observed yet. Either way the record is stale.
	stalePID := entry.pid
	entry.state = StateStarting
	entry.lastExitError = ""
	entry.cmd = nil
	entry.pid = 0
	entry.pgid = 0
	entry.runStartedAt = time.Time{}
	logPath := entry.logPath
	if logPath == "" && s.opts.LogDir != "" {
		logPath = LogPath(s.opts.LogDir, id)
		entry.logPath = logPath
	}
	workingDir := def.WorkingDir
	detached := def.Detached
	s.mu.Unlock()

	if stalePID > 0 {
		_ = s.runtime.Forget(id)
	}

	if err := EnsureWorkingDir(workingDir); err != nil {
		s.recordStartFailure(id, err.Error())
		return err
	}
	if s.opts.Prepare != nil {
		if err := s.opts.Prepare(def); err != nil {
			s.recordStartFailure(id, err.Error())
			return err
		}
	}
	logFile, err := OpenLog(logPath)
	if err != nil {
		s.recordStartFailure(id, err.Error())
		return err
	}
	startedAt := s.now()
	fmt.Fprintf(logFile, "\n[%s] starting service %s\n", startedAt.Format(time.RFC3339), id)

	env, shellErr := s.buildEnv(def)
	if shellErr != nil {
		fmt.Fprintf(logFile, "[%s] interactive shell env unavailable, using process env: %v\n", s.now().Format(time.RFC3339), shellErr)
	}

	cmd := s.command(def.Command, env)
	cmd.Dir = workingDir
	cmd.Env = env
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = procAttr(detached)
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(logFile, "[%s] failed to start: %v\n", s.now().Format(time.RFC3339), err)
		_ = logFile.Close()
		s.recordStartFailure(id, err.Error())
		return err
	}

	pid := cmd.Process.Pid
	pgid := pid
	if gid, gidErr := GroupID(pid); gidErr == nil {
		pgid = gid
	}

	s.mu.Lock()
	entry = s.entries[id]
	entry.cmd = cmd
	entry.pid = pid
	entry.pgid = pgid
	entry.state = StateRunning
	entry.lastStartedAt = startedAt.UTC().Format(time.RFC3339)
	entry.desired = true
	entry.adopted = false
	entry.stopRequested = false
	entry.runStartedAt = startedAt
	entry.nextRestartAt = time.Time{}
	entry.lastExitError = ""
	entry.shellEnvError = ""
	if shellErr != nil {
		entry.shellEnvError = shellErr.Error()
	}
	entry.effectivePath = EnvValue(env, "PATH")
	s.mu.Unlock()

	_ = s.runtime.Record(id, RuntimeRecord{
		PID:       pid,
		PGID:      pgid,
		StartedAt: startedAt.UTC().Format(time.RFC3339),
		Command:   def.Command,
	})

	if s.opts.Started != nil {
		if err := s.opts.Started(def, pid, logFile); err != nil {
			s.logf("[service] service %q started but setup failed: %v", id, err)
			fmt.Fprintf(logFile, "[%s] setup failed: %v\n", s.now().Format(time.RFC3339), err)
			s.setRuntimeError(id, fmt.Sprintf("service is running but setup failed: %v", err))
		}
	}

	go s.waitForExit(id, cmd, logFile)
	return nil
}

// command builds the shell invocation. bash -lc resets PATH to the login
// profile's value before running the command, so the computed PATH is exported
// in the command line itself; relying on cmd.Env alone would drop the layers
// this package just added.
func (s *Supervisor) command(command string, env []string) *exec.Cmd {
	args := append([]string(nil), s.shell...)
	if pathVal := EnvValue(env, "PATH"); pathVal != "" {
		command = fmt.Sprintf("export PATH=%s; %s", ShellQuote(pathVal), command)
	}
	args = append(args, command)
	return exec.Command(args[0], args[1:]...)
}

func (s *Supervisor) buildEnv(def Definition) ([]string, error) {
	if s.opts.Env != nil {
		return s.opts.Env(def)
	}
	return BuildEnv(EnvOptions{
		Extra:           def.ExtraEnv,
		ExtraPath:       def.ExtraPath,
		InheritShellEnv: def.InheritInteractiveShellEnv,
	})
}

func (s *Supervisor) waitForExit(id string, cmd *exec.Cmd, logFile *os.File) {
	err := cmd.Wait()
	if err != nil {
		fmt.Fprintf(logFile, "[%s] service exited with error: %v\n", s.now().Format(time.RFC3339), err)
	} else {
		fmt.Fprintf(logFile, "[%s] service exited\n", s.now().Format(time.RFC3339))
	}
	_ = logFile.Close()

	exitAt := s.now()
	s.mu.Lock()
	entry := s.entries[id]
	if entry == nil || entry.cmd != cmd {
		s.mu.Unlock()
		return
	}
	entry.pid = 0
	entry.pgid = 0
	entry.cmd = nil
	entry.lastExitedAt = exitAt.UTC().Format(time.RFC3339)
	stopRequested := entry.stopRequested
	runDuration := time.Duration(0)
	if !entry.runStartedAt.IsZero() {
		runDuration = exitAt.Sub(entry.runStartedAt)
	}
	entry.runStartedAt = time.Time{}
	if stopRequested {
		entry.state = StateStopped
		entry.lastExitError = ""
		entry.failures = 0
		entry.nextRestartAt = time.Time{}
	} else {
		if runDuration >= StableRunResetWindow {
			entry.failures = 0
		}
		entry.failures++
		entry.nextRestartAt = BackoffTime(entry.failures, exitAt)
		if err != nil {
			entry.lastExitError = err.Error()
		} else {
			entry.lastExitError = fmt.Sprintf("service exited unexpectedly; retrying in %s", entry.nextRestartAt.Sub(exitAt).Round(time.Second))
		}
		entry.state = StateError
	}
	entry.stopRequested = false
	def := entry.def
	s.mu.Unlock()

	_ = s.runtime.Forget(id)
	if s.opts.Exited != nil {
		s.opts.Exited(def)
	}
}

// Stop terminates id and waits for it to exit. Stopping a service that is not
// running is not an error.
func (s *Supervisor) Stop(id string) error { return s.stop(id, true) }

func (s *Supervisor) stop(id string, wait bool) error {
	s.mu.Lock()
	entry, ok := s.entries[id]
	if !ok {
		_, hasDef := s.defs[id]
		s.mu.Unlock()
		if !hasDef {
			return fmt.Errorf("service %q not found", id)
		}
		return nil
	}
	entry.desired = false
	entry.stopRequested = true
	entry.state = StateStopped
	pid := entry.pid
	if pid == 0 {
		entry.cmd = nil
		entry.failures = 0
		entry.nextRestartAt = time.Time{}
		entry.runStartedAt = time.Time{}
	}
	def := entry.def
	s.mu.Unlock()

	if s.opts.Stopping != nil {
		s.opts.Stopping(def)
	}
	if pid == 0 {
		_ = s.runtime.Forget(id)
		return nil
	}
	if !Alive(pid) {
		// Zombies and already-reaped pids still pass kill(pid, 0).
		s.clearPID(id, pid)
		return nil
	}
	if err := StopProcess(pid); err != nil {
		if !Alive(pid) {
			s.clearPID(id, pid)
			return nil
		}
		return err
	}
	if wait {
		WaitForExit(pid, DefaultKillTimeout)
	}
	s.clearPID(id, pid)
	return nil
}

func (s *Supervisor) clearPID(id string, pid int) {
	s.mu.Lock()
	if entry := s.entries[id]; entry != nil && entry.pid == pid {
		entry.pid = 0
		entry.pgid = 0
		entry.cmd = nil
		entry.runStartedAt = time.Time{}
	}
	s.mu.Unlock()
	_ = s.runtime.Forget(id)
}

// Restart stops id and starts it again, ignoring crash backoff.
func (s *Supervisor) Restart(id string) error {
	if err := s.stop(id, true); err != nil {
		return err
	}
	return s.start(id, true)
}

// SetEnabled persists the enabled flag on a definition and starts or stops the
// process to match. Disabling stops the service now.
func (s *Supervisor) SetEnabled(id string, enabled bool) error {
	s.mu.Lock()
	def, ok := s.defs[id]
	if !ok {
		s.mu.Unlock()
		return fmt.Errorf("service %q not found", id)
	}
	def.Enabled = Bool(enabled)
	s.defs[id] = def
	if entry := s.entries[id]; entry != nil {
		entry.def = def
	}
	s.mu.Unlock()
	if enabled {
		return s.start(id, true)
	}
	return s.stop(id, true)
}

// Running reports whether id currently has a live process.
func (s *Supervisor) Running(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.entries[id]
	return entry != nil && entry.pid > 0 && Alive(entry.pid)
}

// Process returns a snapshot of id.
func (s *Supervisor) Process(id string) (Process, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := s.entries[id]
	if entry == nil {
		return Process{}, false
	}
	return snapshotProcess(entry), true
}

// Processes returns a snapshot of every service the supervisor has observed,
// sorted by id.
func (s *Supervisor) Processes() []Process {
	s.mu.Lock()
	defer s.mu.Unlock()
	procs := make([]Process, 0, len(s.entries))
	for _, entry := range s.entries {
		procs = append(procs, snapshotProcess(entry))
	}
	sort.Slice(procs, func(i, j int) bool { return procs[i].Definition.ID < procs[j].Definition.ID })
	return procs
}

func snapshotProcess(entry *serviceProcess) Process {
	return Process{
		Definition:    entry.def,
		State:         entry.state,
		PID:           entry.pid,
		PGID:          entry.pgid,
		Desired:       entry.desired,
		Detached:      entry.def.Detached,
		Adopted:       entry.adopted,
		LogPath:       entry.logPath,
		LastStartedAt: entry.lastStartedAt,
		LastExitedAt:  entry.lastExitedAt,
		LastExitError: entry.lastExitError,
		ShellEnvError: entry.shellEnvError,
		EffectivePath: entry.effectivePath,
		Failures:      entry.failures,
		NextRestartAt: entry.nextRestartAt,
		RunStartedAt:  entry.runStartedAt,
	}
}

func (s *Supervisor) setRuntimeError(id, message string) {
	s.mu.Lock()
	entry := s.entries[id]
	if entry == nil {
		s.mu.Unlock()
		return
	}
	// Keep the pid: the process is still running and must remain stoppable.
	entry.state = StateError
	entry.lastExitError = message
	entry.lastExitedAt = s.now().UTC().Format(time.RFC3339)
	s.mu.Unlock()
}

func (s *Supervisor) recordStartFailure(id, message string) {
	s.mu.Lock()
	entry := s.entries[id]
	if entry == nil {
		s.mu.Unlock()
		return
	}
	entry.pid = 0
	entry.pgid = 0
	entry.cmd = nil
	entry.runStartedAt = time.Time{}
	entry.state = StateError
	entry.lastExitError = message
	now := s.now()
	entry.lastExitedAt = now.UTC().Format(time.RFC3339)
	entry.failures++
	entry.nextRestartAt = BackoffTime(entry.failures, now)
	s.mu.Unlock()
}

// AutoStart starts every enabled definition, in sorted id order. It returns the
// ids it attempted.
func (s *Supervisor) AutoStart() []string {
	defs := s.Definitions()
	started := make([]string, 0, len(defs))
	for _, def := range defs {
		if !def.IsEnabled() {
			continue
		}
		if err := s.start(def.ID, false); err != nil {
			s.logf("[service] autostart failed for %s: %v", def.ID, err)
			continue
		}
		started = append(started, def.ID)
	}
	return started
}

// StartHealthCheck re-reconciles desired services on a ticker. It is
// idempotent; StopHealthCheck or Shutdown ends it.
func (s *Supervisor) StartHealthCheck(interval time.Duration) {
	if interval <= 0 {
		interval = DefaultHealthCheckInterval
	}
	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return
	}
	s.started = true
	stopCh := make(chan struct{})
	s.stopCh = stopCh
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-stopCh:
				return
			case <-ticker.C:
				s.Reconcile()
			}
		}
	}()
}

// StopHealthCheck ends the health check loop. A later StartHealthCheck starts a
// fresh loop, so a host can pause checks and resume them.
func (s *Supervisor) StopHealthCheck() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopCh != nil {
		close(s.stopCh)
		s.stopCh = nil
	}
	s.started = false
}

// Reconcile restarts services that should be running but have no live process
// and whose backoff has elapsed. It returns the ids it attempted.
func (s *Supervisor) Reconcile() []string {
	now := s.now()
	s.mu.Lock()
	var ids []string
	for id, entry := range s.entries {
		if !entry.desired {
			continue
		}
		def, ok := s.defs[id]
		if !ok || !def.IsEnabled() {
			continue
		}
		if entry.state == StateStarting {
			continue
		}
		if entry.pid > 0 && Alive(entry.pid) {
			continue
		}
		if !entry.nextRestartAt.IsZero() && now.Before(entry.nextRestartAt) {
			continue
		}
		ids = append(ids, id)
	}
	s.mu.Unlock()

	sort.Strings(ids)
	for _, id := range ids {
		if err := s.start(id, false); err != nil {
			s.logf("[service] healthcheck restart failed for %s: %v", id, err)
		}
	}
	return ids
}

// Shutdown stops every service it started that is not detached, concurrently.
// Detached services are left running on purpose: the user asked for them to
// outlive the host.
func (s *Supervisor) Shutdown() {
	s.StopHealthCheck()
	s.mu.Lock()
	var ids []string
	for id, entry := range s.entries {
		if entry.def.Detached {
			continue
		}
		if entry.pid > 0 || entry.desired {
			ids = append(ids, id)
		}
	}
	s.mu.Unlock()

	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			_ = s.stop(id, true)
		}(id)
	}
	wg.Wait()
}

// ForceStopAll stops every service, detached ones included.
func (s *Supervisor) ForceStopAll() {
	s.StopHealthCheck()
	s.mu.Lock()
	ids := make([]string, 0, len(s.entries))
	for id := range s.entries {
		ids = append(ids, id)
	}
	s.mu.Unlock()
	for _, id := range ids {
		if err := s.stop(id, true); err != nil {
			s.logf("[service] failed to stop %s: %v", id, err)
		}
	}
}

// ReclaimPolicy decides what happens to a process a previous host instance
// recorded and left running.
type ReclaimPolicy string

const (
	// ReclaimStop kills the orphan so a fresh start can bind its port. This is
	// the right policy when the host owns the process lifetime (ai-critic).
	ReclaimStop ReclaimPolicy = "stop"
	// ReclaimAdopt keeps the orphan and reports it as running, so a host whose
	// services must survive restarts (eluc's standalone llm-proxy) does not
	// interrupt them just because it restarted.
	ReclaimAdopt ReclaimPolicy = "adopt"
)

// ReclaimResult reports what Reclaim did, for logging at boot.
type ReclaimResult struct {
	Reclaimed []string
	Adopted   []string
	// Skipped lists ids whose recorded pid is gone or now runs a different
	// command; their records were discarded.
	Skipped []string
}

// Reclaim reconciles the runtime file against live processes.
//
// A record is trusted only when its pid is alive and its command line still
// contains the recorded command, so a recycled pid is never killed or adopted.
func (s *Supervisor) Reclaim(policy ReclaimPolicy) ReclaimResult {
	result := ReclaimResult{}
	if s.runtime == nil {
		return result
	}
	state := s.runtime.Load()
	for _, id := range state.IDs() {
		rec := state.Processes[id]
		if rec.PID <= 0 {
			_ = s.runtime.Forget(id)
			result.Skipped = append(result.Skipped, id)
			continue
		}
		if !Alive(rec.PID) {
			_ = s.runtime.Forget(id)
			result.Skipped = append(result.Skipped, id)
			continue
		}
		if !CommandMatches(rec.Command, ProcessCommandLine(rec.PID)) {
			s.logf("[service] not reclaiming %s: pid %d no longer runs the recorded command", id, rec.PID)
			_ = s.runtime.Forget(id)
			result.Skipped = append(result.Skipped, id)
			continue
		}
		if policy == ReclaimAdopt {
			s.adopt(id, rec)
			result.Adopted = append(result.Adopted, id)
			continue
		}
		s.logf("[service] reclaiming orphaned process for %q (pid %d) left by a previous instance", id, rec.PID)
		KillProcessGroupNow(rec.PID)
		s.mu.Lock()
		if entry := s.entries[id]; entry != nil && entry.pid == rec.PID {
			entry.pid = 0
			entry.pgid = 0
		}
		s.mu.Unlock()
		_ = s.runtime.Forget(id)
		result.Reclaimed = append(result.Reclaimed, id)
	}
	return result
}

// adopt records a live process this supervisor did not start. There is no
// waiter: Reconcile notices when it dies and starts a replacement.
func (s *Supervisor) adopt(id string, rec RuntimeRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry := s.entries[id]; entry != nil && entry.pid > 0 && entry.pid != rec.PID {
		// A process we started ourselves wins over the record on disk.
		return
	}
	entry := s.entries[id]
	if entry == nil {
		entry = &serviceProcess{state: StateStopped, logPath: LogPath(s.opts.LogDir, id)}
		if def, ok := s.defs[id]; ok {
			entry.def = def
		}
		s.entries[id] = entry
	}
	entry.pid = rec.PID
	entry.pgid = rec.PGID
	entry.state = StateRunning
	entry.adopted = true
	entry.desired = true
	entry.lastStartedAt = rec.StartedAt
	entry.runStartedAt = ParseRecordTime(rec.StartedAt)
	entry.lastExitError = ""
	entry.failures = 0
	entry.nextRestartAt = time.Time{}
}

func removeString(values []string, target string) []string {
	out := values[:0]
	for _, value := range values {
		if value != target {
			out = append(out, value)
		}
	}
	return out
}
