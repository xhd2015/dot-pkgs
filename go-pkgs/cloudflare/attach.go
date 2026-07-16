package cloudflare

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// AttachOptions configures Attach for a hostname on a managed named tunnel.
type AttachOptions struct {
	Domain     string // required public hostname
	LocalURL   string // e.g. http://127.0.0.1:6321; default http://127.0.0.1:6321
	TunnelName string // empty → DefaultTunnelName
	ConfigDir  string // empty → DefaultConfigDir(); tests use t.TempDir()
	Log        io.Writer
	Runner     CommandRunner // nil = real cloudflared; non-nil = fake (tests)
	DNSDeleter DNSDeleter    // optional; carried on Session for best-effort Stop DNS cleanup
	OwnerPID   int           // optional; default os.Getpid()
}

// Attach merges hostname into the managed tunnel registry and ensures a connector.
// Under exclusive flock on the managed dir lock file it ensures the tunnel,
// routes DNS, updates state.json Hosts, writes config.yml, and starts/restarts
// the connector when this is the first attach or ingress changed.
func Attach(opts AttachOptions) (*Session, error) {
	if strings.TrimSpace(opts.Domain) == "" {
		return nil, fmt.Errorf("domain is required")
	}

	localURL := opts.LocalURL
	if localURL == "" {
		localURL = "http://127.0.0.1:6321"
	}
	name := opts.TunnelName
	if name == "" {
		name = DefaultTunnelName
	}
	configDir := opts.ConfigDir
	if configDir == "" {
		dir, err := DefaultConfigDir()
		if err != nil {
			return nil, err
		}
		configDir = dir
	}
	logw := opts.Log
	if logw == nil {
		logw = io.Discard
	}
	ownerPID := opts.OwnerPID
	if ownerPID == 0 {
		ownerPID = os.Getpid()
	}
	runner := opts.Runner

	managedDir, err := ManagedTunnelDir(configDir, name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(managedDir, 0o755); err != nil {
		return nil, fmt.Errorf("create managed tunnel dir: %w", err)
	}

	unlock, err := flockManagedDir(managedDir)
	if err != nil {
		return nil, err
	}
	defer unlock()

	resolvedName, tunnelID, credFile, err := ensureTunnelForSession(runner, name)
	if err != nil {
		return nil, err
	}
	name = resolvedName

	if err := RouteDNS(runner, name, opts.Domain); err != nil {
		return nil, err
	}

	st, err := LoadTunnelState(managedDir)
	if err != nil {
		return nil, err
	}
	if st.Hosts == nil {
		st.Hosts = map[string]*HostEntry{}
	}

	prev, hadHost := st.Hosts[opts.Domain]
	ingressChanged := !hadHost || prev == nil || prev.Service != localURL

	st.TunnelName = name
	st.TunnelID = tunnelID
	st.CredentialsFile = credFile
	st.Hosts[opts.Domain] = &HostEntry{
		Service:    localURL,
		OwnerPID:   ownerPID,
		AttachedAt: time.Now().UTC().Format(time.RFC3339),
	}

	cfg, err := BuildConfigFromState(st)
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(managedDir, "config.yml")
	if err := WriteConfig(configPath, cfg); err != nil {
		return nil, err
	}

	sess := &Session{
		TunnelName: name,
		Domain:     opts.Domain,
		ConfigPath: configPath,
		TunnelID:   tunnelID,
		CredFile:   credFile,
		runner:     runner,
		dnsDeleter: opts.DNSDeleter,
		log:        logw,
		publicURL:  "https://" + opts.Domain,
		managed:    true,
		configDir:  configDir,
	}

	needStart := false
	if runner != nil {
		// Fake mode: PID>0 means a prior successful run for this registry.
		// Start on first attach (no prior run) or when ingress changed.
		if st.ConnectorPID <= 0 || ingressChanged {
			needStart = true
		}
	} else {
		alive := processAlive(st.ConnectorPID)
		if !alive {
			needStart = true
		} else if ingressChanged {
			stopPID(st.ConnectorPID)
			st.ConnectorPID = 0
			needStart = true
		}
	}

	if needStart {
		if runner != nil {
			if _, err := runner.Exec("cloudflared", "tunnel", "--config", configPath, "run", name); err != nil {
				if _, err2 := runner.Exec("cloudflared", "tunnel", "run", "--config", configPath, name); err2 != nil {
					return nil, fmt.Errorf("start cloudflared via runner: %v / %v", err, err2)
				}
			}
			sess.runnerMode = true
			// Sentinel PID so subsequent same-host same-URL attaches skip re-run.
			if st.ConnectorPID <= 0 {
				st.ConnectorPID = 1
			}
		} else {
			logPath := filepath.Join(managedDir, "cloudflared.log")
			logFile, lerr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
			if lerr != nil {
				return nil, lerr
			}
			proc, perr := StartProcess(configPath, name, logFile)
			_ = logFile.Close()
			if perr != nil {
				return nil, perr
			}
			sess.proc = proc
			st.ConnectorPID = proc.PID()
		}
	} else if runner != nil {
		sess.runnerMode = true
	}

	if err := SaveTunnelState(managedDir, st); err != nil {
		return nil, err
	}

	fmt.Fprintf(logw, "cloudflare attach: tunnel=%s domain=%s url=%s\n", name, opts.Domain, sess.publicURL)
	return sess, nil
}

// flockManagedDir takes an exclusive flock on {dir}/lock. The returned function
// unlocks and closes the lock file.
func flockManagedDir(dir string) (unlock func(), err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create managed dir for lock: %w", err)
	}
	lockPath := filepath.Join(dir, "lock")
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open managed lock: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("flock managed lock: %w", err)
	}
	return func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		_ = f.Close()
	}, nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil
}

func stopPID(pid int) {
	if pid <= 0 {
		return
	}
	// Prefer process-group kill (StartProcess uses Setpgid).
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	_ = syscall.Kill(pid, syscall.SIGTERM)
}
