package cloudflare

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// DetachOptions configures Detach for releasing one hostname from a managed tunnel.
type DetachOptions struct {
	Domain     string
	TunnelName string
	ConfigDir  string
	Log        io.Writer
	Runner     CommandRunner
	DNSDeleter DNSDeleter // optional; best-effort
}

// Detach removes one hostname from the managed tunnel registry.
// Missing hosts are a no-op success. DNS delete failures are logged and do not
// fail Detach. When hosts remain, the connector is restarted with the rewritten
// config; when Hosts is empty, the connector is stopped and ConnectorPID is 0.
func Detach(opts DetachOptions) error {
	domain := strings.TrimSpace(opts.Domain)
	if domain == "" {
		return fmt.Errorf("domain is required")
	}

	name := opts.TunnelName
	if name == "" {
		name = DefaultTunnelName
	}
	configDir := opts.ConfigDir
	if configDir == "" {
		dir, err := DefaultConfigDir()
		if err != nil {
			return err
		}
		configDir = dir
	}
	logw := opts.Log
	if logw == nil {
		logw = io.Discard
	}
	runner := opts.Runner

	managedDir, err := ManagedTunnelDir(configDir, name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(managedDir, 0o755); err != nil {
		return fmt.Errorf("create managed tunnel dir: %w", err)
	}

	unlock, err := flockManagedDir(managedDir)
	if err != nil {
		return err
	}
	defer unlock()

	st, err := LoadTunnelState(managedDir)
	if err != nil {
		return err
	}
	if st.Hosts == nil {
		st.Hosts = map[string]*HostEntry{}
	}

	if _, hadHost := st.Hosts[domain]; !hadHost {
		// Prefer no-op success when domain is not in the registry.
		return nil
	}

	delete(st.Hosts, domain)

	// Preserve tunnel identity fields for config rewrite.
	if st.TunnelName == "" {
		st.TunnelName = name
	}

	cfg, err := BuildConfigFromState(st)
	if err != nil {
		return err
	}
	configPath := filepath.Join(managedDir, "config.yml")
	if err := WriteConfig(configPath, cfg); err != nil {
		return err
	}

	if len(st.Hosts) == 0 {
		// Last host: stop connector and clear logical PID.
		if runner == nil {
			stopPID(st.ConnectorPID)
		}
		// runnerMode has no OS process; clearing PID marks connector down.
		st.ConnectorPID = 0
	} else {
		// Siblings remain: restart connector with remaining hosts.
		if err := restartManagedConnector(st, runner, configPath, name, managedDir); err != nil {
			return err
		}
	}

	if err := SaveTunnelState(managedDir, st); err != nil {
		return err
	}

	// Best-effort DNS cleanup; never fail Detach for DNS-only errors.
	if opts.DNSDeleter != nil {
		if err := DeleteDNS(opts.DNSDeleter, domain); err != nil {
			fmt.Fprintf(logw, "warning: DeleteDNS(%s): %v\n", domain, err)
		}
	}

	fmt.Fprintf(logw, "cloudflare detach: tunnel=%s domain=%s hosts_left=%d\n", name, domain, len(st.Hosts))
	return nil
}

// restartManagedConnector restarts the cloudflared connector after ingress change.
func restartManagedConnector(st *TunnelState, runner CommandRunner, configPath, name, managedDir string) error {
	if runner != nil {
		if _, err := runner.Exec("cloudflared", "tunnel", "--config", configPath, "run", name); err != nil {
			if _, err2 := runner.Exec("cloudflared", "tunnel", "run", "--config", configPath, name); err2 != nil {
				return fmt.Errorf("restart cloudflared via runner: %v / %v", err, err2)
			}
		}
		if st.ConnectorPID <= 0 {
			st.ConnectorPID = 1
		}
		return nil
	}

	// Real process: stop previous then start a new one.
	stopPID(st.ConnectorPID)
	st.ConnectorPID = 0

	logPath := filepath.Join(managedDir, "cloudflared.log")
	logFile, lerr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if lerr != nil {
		return lerr
	}
	proc, perr := StartProcess(configPath, name, logFile)
	_ = logFile.Close()
	if perr != nil {
		return perr
	}
	st.ConnectorPID = proc.PID()
	return nil
}
