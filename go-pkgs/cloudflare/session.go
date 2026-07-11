package cloudflare

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// StartSession finds/creates a named tunnel, routes DNS, writes config, and starts cloudflared.
func StartSession(opts SessionOptions) (*Session, error) {
	if strings.TrimSpace(opts.Domain) == "" {
		return nil, fmt.Errorf("domain is required")
	}
	localURL := opts.LocalURL
	if localURL == "" {
		localURL = "http://127.0.0.1:6321"
	}
	logw := opts.Log
	if logw == nil {
		logw = io.Discard
	}
	runner := opts.Runner

	workDir := opts.WorkDir
	ownWorkDir := false
	if workDir == "" {
		dir, err := os.MkdirTemp("", "spl-seatalk-local-bot-*")
		if err != nil {
			return nil, fmt.Errorf("create workdir: %w", err)
		}
		workDir = dir
		ownWorkDir = true
	} else if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("create workdir: %w", err)
	}

	name, tunnelID, credFile, err := ensureTunnelForSession(runner, opts.TunnelName)
	if err != nil {
		if ownWorkDir {
			_ = os.RemoveAll(workDir)
		}
		return nil, err
	}

	if err := RouteDNS(runner, name, opts.Domain); err != nil {
		if ownWorkDir {
			_ = os.RemoveAll(workDir)
		}
		return nil, err
	}

	configPath := filepath.Join(workDir, "config.yml")
	cfg := &Config{
		Tunnel:          tunnelID,
		CredentialsFile: credFile,
		Ingress: []IngressRule{
			{Hostname: opts.Domain, Service: localURL},
			{Service: "http_status:404"},
		},
	}
	if err := WriteConfig(configPath, cfg); err != nil {
		if ownWorkDir {
			_ = os.RemoveAll(workDir)
		}
		return nil, err
	}

	sess := &Session{
		TunnelName: name,
		Domain:     opts.Domain,
		WorkDir:    workDir,
		ConfigPath: configPath,
		TunnelID:   tunnelID,
		CredFile:   credFile,
		ownWorkDir: ownWorkDir,
		runner:     runner,
		dnsDeleter: opts.DNSDeleter,
		log:        logw,
		publicURL:  "https://" + opts.Domain,
	}

	// Start cloudflared: injectable runner (tests) or real process.
	if runner != nil {
		// Fake runners return immediately for `tunnel ... run`.
		if _, err := runner.Exec("cloudflared", "tunnel", "--config", configPath, "run", name); err != nil {
			// Some fakes may only match args containing "run"; try alternate order used by cloudflared docs.
			if _, err2 := runner.Exec("cloudflared", "tunnel", "run", "--config", configPath, name); err2 != nil {
				_ = sess.cleanupPartial()
				return nil, fmt.Errorf("start cloudflared via runner: %v / %v", err, err2)
			}
		}
		sess.runnerMode = true
	} else {
		logPath := filepath.Join(workDir, "cloudflared.log")
		logFile, lerr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if lerr != nil {
			_ = sess.cleanupPartial()
			return nil, lerr
		}
		proc, perr := StartProcess(configPath, name, logFile)
		if perr != nil {
			_ = logFile.Close()
			_ = sess.cleanupPartial()
			return nil, perr
		}
		sess.proc = proc
		// log file lives for process lifetime; process inherits fd — close our handle is OK after start
		_ = logFile.Close()
	}

	fmt.Fprintf(logw, "cloudflare session started: tunnel=%s domain=%s url=%s\n", name, opts.Domain, sess.publicURL)
	return sess, nil
}

// PublicBaseURL returns https://<domain> (no path).
func (s *Session) PublicBaseURL() string {
	if s == nil {
		return ""
	}
	return s.publicURL
}

// Stop stops cloudflared, removes the session workdir, and best-effort deletes DNS.
func (s *Session) Stop() error {
	if s == nil {
		return nil
	}
	var firstErr error

	if s.proc != nil {
		if err := s.proc.Stop(); err != nil && firstErr == nil {
			firstErr = err
		}
		s.proc = nil
	}
	// runnerMode: nothing OS-level to kill

	if s.WorkDir != "" {
		if err := os.RemoveAll(s.WorkDir); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if s.dnsDeleter != nil && s.Domain != "" {
		if err := DeleteDNS(s.dnsDeleter, s.Domain); err != nil {
			fmt.Fprintf(logOrDiscard(s.log), "warning: DeleteDNS(%s): %v\n", s.Domain, err)
			// best-effort: do not fail Stop for DNS delete errors (SIGINT path stays clean)
		}
	}

	return firstErr
}

func (s *Session) cleanupPartial() error {
	if s == nil {
		return nil
	}
	if s.ownWorkDir && s.WorkDir != "" {
		return os.RemoveAll(s.WorkDir)
	}
	return nil
}

func logOrDiscard(w io.Writer) io.Writer {
	if w == nil {
		return io.Discard
	}
	return w
}
