package cloudflare

import (
	"strings"
	"testing"
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

func TestTunnelRunArgsPinsProtocol(t *testing.T) {
	got := tunnelRunArgs("/tmp/managed/config.yml", "tunnel-a")
	if v := findFlag(got, "--config"); v != "/tmp/managed/config.yml" {
		t.Fatalf("--config = %q", v)
	}
	if v := findFlag(got, "--protocol"); v != DefaultTunnelProtocol {
		t.Fatalf("--protocol = %q, want %q", v, DefaultTunnelProtocol)
	}
	if DefaultTunnelProtocol != "http2" {
		t.Fatalf("DefaultTunnelProtocol = %q, want http2", DefaultTunnelProtocol)
	}
	if n := len(got); n < 2 || got[n-2] != "run" || got[n-1] != "tunnel-a" {
		t.Fatalf("argv must end with `run <name>`, got %v", got)
	}
	if strings.Join(got[:2], " ") != "tunnel --config" {
		t.Fatalf("argv must start with `tunnel --config`, got %v", got)
	}
}

func TestTunnelRunArgsAltPinsProtocol(t *testing.T) {
	got := tunnelRunArgsAlt("/tmp/managed/config.yml", "tunnel-a")
	if v := findFlag(got, "--protocol"); v != DefaultTunnelProtocol {
		t.Fatalf("--protocol = %q, want %q", v, DefaultTunnelProtocol)
	}
	if v := findFlag(got, "--config"); v != "/tmp/managed/config.yml" {
		t.Fatalf("--config = %q", v)
	}
	// `run` precedes the tunnel name in the fallback shape.
	runIdx := -1
	for i, arg := range got {
		if arg == "run" {
			runIdx = i
			break
		}
	}
	if runIdx < 0 || got[len(got)-1] != "tunnel-a" {
		t.Fatalf("fallback shape must be `tunnel run --config <cfg> --protocol http2 <name>`, got %v", got)
	}
}

// TestDefaultStartCommandPinsProtocol guards the real spawn argv (not the
// test-injected hook): every registry-started connector must carry the pin.
func TestDefaultStartCommandPinsProtocol(t *testing.T) {
	cmd := startCommand("/tmp/managed/config.yml", "tunnel-a")
	if cmd.Args[0] != "cloudflared" {
		t.Fatalf("binary = %q", cmd.Args[0])
	}
	if v := findFlag(cmd.Args, "--protocol"); v != "http2" {
		t.Fatalf("default startCommand --protocol = %q, want http2; argv: %v", v, cmd.Args)
	}
}

// TestConfigFromArgsPinnedShape ensures arg-based connector matching
// (stale-connector sweep) still resolves --config on the pinned argv.
func TestConfigFromArgsPinnedShape(t *testing.T) {
	got := configFromArgs(tunnelRunArgs("/tmp/managed/config.yml", "tunnel-a"))
	if got != "/tmp/managed/config.yml" {
		t.Fatalf("configFromArgs on pinned argv = %q", got)
	}
}
