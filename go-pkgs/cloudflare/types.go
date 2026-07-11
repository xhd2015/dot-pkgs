package cloudflare

import "io"

// DefaultTunnelName is the stable named tunnel used by local-bot when none is specified.
const DefaultTunnelName = "spl-seatalk-local-bot"

// TunnelInfo represents a Cloudflare tunnel as returned by `cloudflared tunnel list --output json`.
type TunnelInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	CreatedAt   string `json:"created_at,omitempty"`
	Connections []any  `json:"connections,omitempty"`
}

// Config is the cloudflared config.yml structure.
type Config struct {
	Tunnel          string        `yaml:"tunnel"`
	CredentialsFile string        `yaml:"credentials-file"`
	Ingress         []IngressRule `yaml:"ingress"`
}

// IngressRule is a single cloudflared ingress entry.
type IngressRule struct {
	Hostname string `yaml:"hostname,omitempty"`
	Service  string `yaml:"service"`
}

// CommandRunner runs external commands (cloudflared). Method is Exec (not Run)
// so doctest harnesses can keep a package-level Run helper without collisions.
type CommandRunner interface {
	Exec(name string, args ...string) ([]byte, error)
}

// DNSDeleter deletes a DNS hostname (best-effort cleanup).
type DNSDeleter interface {
	DeleteHostname(hostname string) error
}

// SessionOptions configures StartSession.
type SessionOptions struct {
	Domain     string // required public hostname
	LocalURL   string // e.g. http://127.0.0.1:6321
	TunnelName string // empty → DefaultTunnelName
	WorkDir    string // empty → temp dir
	Log        io.Writer
	Runner     CommandRunner
	DNSDeleter DNSDeleter
}

// Session is a running named-tunnel session.
type Session struct {
	TunnelName  string
	Domain      string
	WorkDir     string
	ConfigPath  string
	TunnelID    string
	CredFile    string
	ownWorkDir  bool
	proc        *Process
	runner      CommandRunner
	dnsDeleter  DNSDeleter
	log         io.Writer
	publicURL   string
	runnerMode  bool // process was "started" via Runner.Exec rather than real OS process
}

// StatusInfo reports cloudflared install / auth state.
type StatusInfo struct {
	Installed bool
	Detail    string
}
