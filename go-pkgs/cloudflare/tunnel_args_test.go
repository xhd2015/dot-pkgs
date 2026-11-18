package cloudflare

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// findFlag returns the value following a --flag pair, or "" when absent.
func findFlag(args []string, flag string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == flag {
			return args[i+1]
		}
	}
	return ""
}

func TestTunnelRunArgsCarriesProtocol(t *testing.T) {
	cases := []struct {
		name     string
		protocol string
		want     string
	}{
		{"explicit http2", ProtocolHTTP2, ProtocolHTTP2},
		{"explicit quic", ProtocolQUIC, ProtocolQUIC},
		{"empty falls back to default", "", DefaultTunnelProtocol},
		{"whitespace falls back to default", "  ", DefaultTunnelProtocol},
		{"unknown falls back to default", "tcp", DefaultTunnelProtocol},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := tunnelRunArgs("/tmp/managed/config.yml", "tunnel-a", c.protocol)
			if v := findFlag(got, "--config"); v != "/tmp/managed/config.yml" {
				t.Fatalf("--config = %q", v)
			}
			if v := findFlag(got, "--protocol"); v != c.want {
				t.Fatalf("--protocol = %q, want %q; argv: %v", v, c.want, got)
			}
			if n := len(got); n < 2 || got[n-2] != "run" || got[n-1] != "tunnel-a" {
				t.Fatalf("argv must end with `run <name>`, got %v", got)
			}
			if strings.Join(got[:2], " ") != "tunnel --config" {
				t.Fatalf("argv must start with `tunnel --config`, got %v", got)
			}
		})
	}
}

func TestTunnelRunArgsAltCarriesProtocol(t *testing.T) {
	got := tunnelRunArgsAlt("/tmp/managed/config.yml", "tunnel-a", ProtocolQUIC)
	if v := findFlag(got, "--protocol"); v != ProtocolQUIC {
		t.Fatalf("--protocol = %q, want quic", v)
	}
	if v := findFlag(got, "--config"); v != "/tmp/managed/config.yml" {
		t.Fatalf("--config = %q", v)
	}
	runIdx := -1
	for i, arg := range got {
		if arg == "run" {
			runIdx = i
			break
		}
	}
	if runIdx < 0 || got[len(got)-1] != "tunnel-a" {
		t.Fatalf("fallback shape must end with `run <name>`, got %v", got)
	}
}

// The historical default must stay http2: an install that never configures a
// protocol has to keep the exact pre-existing behaviour.
func TestDefaultTunnelProtocolStaysHTTP2(t *testing.T) {
	if DefaultTunnelProtocol != ProtocolHTTP2 {
		t.Fatalf("DefaultTunnelProtocol = %q, want http2", DefaultTunnelProtocol)
	}
	if got := tunnelRunArgs("/tmp/c.yml", "t", ""); findFlag(got, "--protocol") != ProtocolHTTP2 {
		t.Fatalf("empty protocol did not resolve to http2: %v", got)
	}
}

func TestNormalizeProtocol(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"", DefaultTunnelProtocol, false},
		{"  ", DefaultTunnelProtocol, false},
		{"http2", ProtocolHTTP2, false},
		{"HTTP2", ProtocolHTTP2, false},
		{" quic ", ProtocolQUIC, false},
		{"Quic", ProtocolQUIC, false},
		{"tcp", "", true},
		{"auto", "", true},
	}
	for _, c := range cases {
		got, err := NormalizeProtocol(c.in)
		if c.wantErr {
			if err == nil {
				t.Fatalf("NormalizeProtocol(%q) = %q, want error", c.in, got)
			}
			continue
		}
		if err != nil {
			t.Fatalf("NormalizeProtocol(%q) unexpected error: %v", c.in, err)
		}
		if got != c.want {
			t.Fatalf("NormalizeProtocol(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// The real spawn argv (not the test hook) must carry the protocol as a flag,
// never through the inherited environment.
func TestDefaultStartCommandCarriesProtocol(t *testing.T) {
	t.Setenv("TUNNEL_TRANSPORT_PROTOCOL", ProtocolHTTP2)

	cmd := startCommand("/tmp/managed/config.yml", "tunnel-a", ProtocolQUIC)
	if cmd.Args[0] != "cloudflared" {
		t.Fatalf("binary = %q", cmd.Args[0])
	}
	if v := findFlag(cmd.Args, "--protocol"); v != ProtocolQUIC {
		t.Fatalf("--protocol = %q, want quic; argv: %v", v, cmd.Args)
	}
	// A stray environment value must not influence the argv: the flag decides.
	for _, a := range cmd.Args {
		if strings.Contains(a, "TUNNEL_TRANSPORT_PROTOCOL") {
			t.Fatalf("argv leaked the environment variable: %v", cmd.Args)
		}
	}
}

// A respawn after a crash must reuse the protocol captured at StartProcess
// time; supervise() has no caller to ask.
func TestSuperviseRespawnKeepsProtocol(t *testing.T) {
	var seen []string
	prev := startCommand
	startCommand = func(configPath, tunnelName, protocol string) *exec.Cmd {
		seen = append(seen, protocol)
		return exec.Command("sh", "-c", "exit 0")
	}
	t.Cleanup(func() { startCommand = prev })

	p, err := StartProcess("cfg", "tunnel", ProtocolQUIC, nil)
	if err != nil {
		t.Fatal(err)
	}
	p.sleepFor = func(time.Duration) {}

	deadline := time.Now().Add(5 * time.Second)
	for len(seen) < 3 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	_ = p.Stop()

	if len(seen) < 2 {
		t.Fatalf("expected at least 2 spawns, got %d (%v)", len(seen), seen)
	}
	for i, got := range seen {
		if got != ProtocolQUIC {
			t.Fatalf("spawn %d protocol = %q, want quic", i, got)
		}
	}
}

// A state.json written before Protocol existed must read back as empty and
// resolve to the default, not fail.
func TestTunnelStateProtocolRoundTrip(t *testing.T) {
	dir := t.TempDir()
	// LoadTunnelState/SaveTunnelState take the managed dir, not the file.
	path := filepath.Join(dir, stateFileName)

	legacy := `{"tunnel_name":"t1","hosts":{"a.example":{"service":"http://127.0.0.1:1"}}}`
	if err := os.WriteFile(path, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}
	st, err := LoadTunnelState(dir)
	if err != nil {
		t.Fatalf("LoadTunnelState(legacy): %v", err)
	}
	if st.Protocol != "" {
		t.Fatalf("legacy Protocol = %q, want empty", st.Protocol)
	}
	if got := effectiveProtocol(st.Protocol); got != ProtocolHTTP2 {
		t.Fatalf("legacy effective protocol = %q, want http2", got)
	}

	st.Protocol = ProtocolQUIC
	if err := SaveTunnelState(dir, st); err != nil {
		t.Fatalf("SaveTunnelState: %v", err)
	}
	back, err := LoadTunnelState(dir)
	if err != nil {
		t.Fatalf("LoadTunnelState(round trip): %v", err)
	}
	if back.Protocol != ProtocolQUIC {
		t.Fatalf("round-trip Protocol = %q, want quic", back.Protocol)
	}
}

// arg-based connector matching (stale-connector sweep) must still resolve
// --config with --protocol present.
func TestConfigFromArgsWithProtocol(t *testing.T) {
	got := configFromArgs(tunnelRunArgs("/tmp/managed/config.yml", "tunnel-a", ProtocolQUIC))
	if got != "/tmp/managed/config.yml" {
		t.Fatalf("configFromArgs on protocol argv = %q", got)
	}
}
