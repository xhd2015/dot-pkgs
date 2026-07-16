package cloudflare

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

const (
	managedTunnelsSegment = "managed-tunnels"
	stateFileName         = "state.json"
)

// TunnelState is the on-disk state.json for a managed tunnel directory.
type TunnelState struct {
	TunnelName      string                `json:"tunnel_name"`
	TunnelID        string                `json:"tunnel_id,omitempty"`
	CredentialsFile string                `json:"credentials_file,omitempty"`
	ConnectorPID    int                   `json:"connector_pid,omitempty"`
	Hosts           map[string]*HostEntry `json:"hosts"` // key = hostname
}

// HostEntry is one hostname attached to a managed tunnel.
type HostEntry struct {
	Service    string `json:"service"`
	OwnerPID   int    `json:"owner_pid,omitempty"`
	AttachedAt string `json:"attached_at,omitempty"` // RFC3339 optional
}

// ManagedTunnelsRoot returns {configDir}/managed-tunnels.
func ManagedTunnelsRoot(configDir string) string {
	return filepath.Join(configDir, managedTunnelsSegment)
}

// TunnelNameSafe maps a human tunnel name to a filesystem-safe directory segment.
// Normal [a-zA-Z0-9_-] names are lowercased; path-unsafe and other non-safe
// characters are replaced with '-'.
func TunnelNameSafe(tunnelName string) string {
	if tunnelName == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(tunnelName))
	for _, r := range tunnelName {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(unicode.ToLower(r))
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-':
			b.WriteRune(r)
		default:
			// Path separators and other unsafe characters become '-'.
			b.WriteByte('-')
		}
	}
	return b.String()
}

// ManagedTunnelDir resolves {configDir}/managed-tunnels/<TunnelNameSafe(name)>.
// Empty tunnel names are rejected.
func ManagedTunnelDir(configDir, tunnelName string) (string, error) {
	if strings.TrimSpace(tunnelName) == "" {
		return "", fmt.Errorf("empty tunnel name")
	}
	safe := TunnelNameSafe(tunnelName)
	if safe == "" {
		return "", fmt.Errorf("empty tunnel name")
	}
	return filepath.Join(ManagedTunnelsRoot(configDir), safe), nil
}

// LoadTunnelState reads state.json under dir. A missing file returns an empty
// state with a non-nil Hosts map and a nil error (attach-friendly).
func LoadTunnelState(dir string) (*TunnelState, error) {
	path := filepath.Join(dir, stateFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TunnelState{Hosts: map[string]*HostEntry{}}, nil
		}
		return nil, fmt.Errorf("read tunnel state %s: %w", path, err)
	}
	var st TunnelState
	if err := json.Unmarshal(data, &st); err != nil {
		return nil, fmt.Errorf("parse tunnel state %s: %w", path, err)
	}
	if st.Hosts == nil {
		st.Hosts = map[string]*HostEntry{}
	}
	return &st, nil
}

// SaveTunnelState creates dir if needed and writes state.json.
func SaveTunnelState(dir string, st *TunnelState) error {
	if st == nil {
		return fmt.Errorf("tunnel state is nil")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create tunnel dir %s: %w", dir, err)
	}
	// Ensure Hosts serializes as {} rather than null.
	if st.Hosts == nil {
		st.Hosts = map[string]*HostEntry{}
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal tunnel state: %w", err)
	}
	data = append(data, '\n')
	path := filepath.Join(dir, stateFileName)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write tunnel state %s: %w", path, err)
	}
	return nil
}

// BuildConfigFromState builds a cloudflared Config from TunnelState.
// Ingress rules are one per host (hostnames sorted for determinism), then a
// catch-all {Service: "http_status:404"} with empty Hostname.
// Tunnel and CredentialsFile are copied from state (Tunnel = TunnelID).
func BuildConfigFromState(st *TunnelState) (*Config, error) {
	if st == nil {
		return nil, fmt.Errorf("tunnel state is nil")
	}
	hosts := st.Hosts
	if hosts == nil {
		hosts = map[string]*HostEntry{}
	}
	keys := make([]string, 0, len(hosts))
	for h := range hosts {
		keys = append(keys, h)
	}
	sort.Strings(keys)

	ingress := make([]IngressRule, 0, len(keys)+1)
	for _, h := range keys {
		service := ""
		if e := hosts[h]; e != nil {
			service = e.Service
		}
		ingress = append(ingress, IngressRule{
			Hostname: h,
			Service:  service,
		})
	}
	ingress = append(ingress, IngressRule{
		Service: "http_status:404",
	})

	return &Config{
		Tunnel:          st.TunnelID,
		CredentialsFile: st.CredentialsFile,
		Ingress:         ingress,
	}, nil
}
