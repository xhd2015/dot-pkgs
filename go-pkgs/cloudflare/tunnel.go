package cloudflare

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ListTunnels runs `cloudflared tunnel list --output json`.
func ListTunnels(runner CommandRunner) ([]TunnelInfo, error) {
	out, err := execCloudflared(runner, "tunnel", "list", "--output", "json")
	if err != nil {
		return nil, fmtCmdErr("tunnel list", out, err)
	}
	// cloudflared may print warnings before JSON; try raw then last JSON array.
	data := out
	if idx := indexJSONArray(out); idx > 0 {
		data = out[idx:]
	}
	var tunnels []TunnelInfo
	if err := json.Unmarshal(data, &tunnels); err != nil {
		return nil, fmt.Errorf("tunnel list: parse json: %w", err)
	}
	return tunnels, nil
}

func indexJSONArray(b []byte) int {
	for i, c := range b {
		if c == '[' {
			return i
		}
	}
	return -1
}

// FindOrCreateTunnel lists tunnels and returns preferredName if present; otherwise creates it.
func FindOrCreateTunnel(runner CommandRunner, preferredName string) (string, error) {
	if preferredName == "" {
		preferredName = DefaultTunnelName
	}
	tunnels, err := ListTunnels(runner)
	if err == nil {
		for _, t := range tunnels {
			if t.Name == preferredName {
				return t.Name, nil
			}
		}
	}
	out, err := execCloudflared(runner, "tunnel", "create", preferredName)
	if err != nil {
		return "", fmt.Errorf("failed to create tunnel %q: %s", preferredName, strings.TrimSpace(string(out)))
	}
	_ = out
	return preferredName, nil
}

// FindTunnelIDAndCreds resolves tunnel ID (via `tunnel info`) and credentials file under ~/.cloudflared.
func FindTunnelIDAndCreds(runner CommandRunner, tunnelRef string) (tunnelID string, credFile string, err error) {
	infoOut, err := execCloudflared(runner, "tunnel", "info", tunnelRef)
	if err != nil {
		return "", "", fmt.Errorf("tunnel %q not found: %s", tunnelRef, strings.TrimSpace(string(infoOut)))
	}
	tunnelID = parseUUIDFromInfo(string(infoOut))
	if tunnelID == "" {
		return "", "", fmt.Errorf("could not determine tunnel ID for %q", tunnelRef)
	}
	cfgDir, derr := DefaultConfigDir()
	if derr != nil {
		return "", "", derr
	}
	credFile = filepath.Join(cfgDir, tunnelID+".json")
	if _, statErr := os.Stat(credFile); statErr != nil {
		return "", "", fmt.Errorf("credentials file not found: %s", credFile)
	}
	return tunnelID, credFile, nil
}

// EnsureTunnelExists finds an existing tunnel or creates it; returns ID and credentials path.
func EnsureTunnelExists(runner CommandRunner, tunnelRef string) (tunnelID string, credFile string, err error) {
	tunnelID, credFile, err = FindTunnelIDAndCreds(runner, tunnelRef)
	if err == nil {
		return tunnelID, credFile, nil
	}

	createOut, createErr := execCloudflared(runner, "tunnel", "create", tunnelRef)
	createStr := strings.TrimSpace(string(createOut))
	if createErr != nil {
		// Maybe it already exists; retry find.
		if id2, cred2, err2 := FindTunnelIDAndCreds(runner, tunnelRef); err2 == nil {
			return id2, cred2, nil
		}
		return "", "", fmt.Errorf("failed to create tunnel %q: %s", tunnelRef, createStr)
	}

	tunnelID, credFile = parseCreateOutput(createStr)
	if credFile == "" && tunnelID != "" {
		if cfgDir, derr := DefaultConfigDir(); derr == nil {
			candidate := filepath.Join(cfgDir, tunnelID+".json")
			if _, statErr := os.Stat(candidate); statErr == nil {
				credFile = candidate
			}
		}
	}
	if tunnelID == "" {
		// try info
		if id2, _, err2 := FindTunnelIDAndCreds(runner, tunnelRef); err2 == nil {
			tunnelID = id2
		} else if infoOut, infoErr := execCloudflared(runner, "tunnel", "info", tunnelRef); infoErr == nil {
			tunnelID = parseUUIDFromInfo(string(infoOut))
		}
	}
	if tunnelID == "" {
		return "", "", fmt.Errorf("could not determine tunnel ID for %q", tunnelRef)
	}
	if credFile == "" {
		return "", "", fmt.Errorf("could not find credentials file for tunnel %q (id: %s)", tunnelRef, tunnelID)
	}
	return tunnelID, credFile, nil
}

// ensureTunnelForSession finds/creates tunnel and resolves credentials without double-create races.
// Prefers create-output credentials when the tunnel is missing from list.
func ensureTunnelForSession(runner CommandRunner, preferredName string) (name, tunnelID, credFile string, err error) {
	name = preferredName
	if name == "" {
		name = DefaultTunnelName
	}

	exists := false
	if tunnels, listErr := ListTunnels(runner); listErr == nil {
		for _, t := range tunnels {
			if t.Name == name {
				exists = true
				if t.ID != "" && IsUUID(t.ID) {
					tunnelID = t.ID
				}
				break
			}
		}
	}

	if !exists {
		out, createErr := execCloudflared(runner, "tunnel", "create", name)
		createStr := strings.TrimSpace(string(out))
		if createErr != nil {
			// fall through to resolve existing
			_ = createErr
		} else {
			tunnelID, credFile = parseCreateOutput(createStr)
		}
	}

	if tunnelID == "" || credFile == "" || (credFile != "" && fileMissing(credFile)) {
		if id, cred, ferr := FindTunnelIDAndCreds(runner, name); ferr == nil {
			if tunnelID == "" {
				tunnelID = id
			}
			if credFile == "" || fileMissing(credFile) {
				credFile = cred
			}
		} else if tunnelID == "" {
			// last resort: info only
			if infoOut, infoErr := execCloudflared(runner, "tunnel", "info", name); infoErr == nil {
				tunnelID = parseUUIDFromInfo(string(infoOut))
			}
		}
	}

	if tunnelID == "" {
		return "", "", "", fmt.Errorf("could not determine tunnel ID for %q", name)
	}
	if credFile == "" {
		// allow missing file path default for config write (session tests provide real path from create)
		if cfgDir, derr := DefaultConfigDir(); derr == nil {
			credFile = filepath.Join(cfgDir, tunnelID+".json")
		}
	}
	if credFile == "" {
		return "", "", "", fmt.Errorf("could not find credentials file for tunnel %q", name)
	}
	return name, tunnelID, credFile, nil
}

func fileMissing(path string) bool {
	_, err := os.Stat(path)
	return err != nil
}

// RouteDNS creates a DNS CNAME for hostname → tunnel, treating already-exists as success.
func RouteDNS(runner CommandRunner, tunnel, hostname string) error {
	out, err := execCloudflared(runner, "tunnel", "route", "dns", "--overwrite-dns", tunnel, hostname)
	if err != nil {
		combined := errOutput(out, err)
		if strings.Contains(combined, "already exists") || strings.Contains(combined, "Added CNAME") {
			return nil
		}
		return fmt.Errorf("failed to create DNS route: %s", strings.TrimSpace(combined))
	}
	return nil
}

// DeleteDNS best-effort deletes hostname via the injectable DNS client.
func DeleteDNS(deleter DNSDeleter, hostname string) error {
	if deleter == nil {
		return fmt.Errorf("DNSDeleter is nil")
	}
	return deleter.DeleteHostname(hostname)
}

// Status reports whether cloudflared is installed (and optional detail).
// When runner is non-nil, calls Exec("cloudflared") with no args (look-path style).
// When runner is nil, uses exec.LookPath.
func Status(runner CommandRunner) (*StatusInfo, error) {
	if runner != nil {
		_, err := runner.Exec("cloudflared")
		if err != nil {
			return &StatusInfo{Installed: false, Detail: err.Error()}, nil
		}
		return &StatusInfo{Installed: true, Detail: "cloudflared found"}, nil
	}
	path, err := lookPath("cloudflared")
	if err != nil {
		return &StatusInfo{Installed: false, Detail: err.Error()}, nil
	}
	return &StatusInfo{Installed: true, Detail: path}, nil
}
