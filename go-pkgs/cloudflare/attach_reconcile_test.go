package cloudflare

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type countingAttachRunner struct {
	list []byte
	runs int
}

func (r *countingAttachRunner) Exec(_ string, args ...string) ([]byte, error) {
	cmd := strings.Join(args, " ")
	if strings.Contains(cmd, "tunnel list") {
		return r.list, nil
	}
	if strings.Contains(cmd, "tunnel info") {
		return []byte("ID: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee\n"), nil
	}
	if strings.Contains(cmd, "tunnel --config") && strings.Contains(cmd, " run ") {
		r.runs++
		return nil, nil
	}
	return nil, nil
}

func TestAttachRestartsWhenConfigDriftsOrExplicitlyReconciled(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cf := filepath.Join(home, ".cloudflared")
	if err := os.MkdirAll(cf, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cf, id+".json"), []byte(`{"AccountTag":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}
	list, _ := json.Marshal([]TunnelInfo{{ID: id, Name: "port-test"}})
	runner := &countingAttachRunner{list: list}
	opts := AttachOptions{Domain: "*.example.com", LocalURL: "unix:/tmp/relay.sock", TunnelName: "port-test", ConfigDir: cf, Runner: runner, SkipDNS: true}
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if runner.runs != 1 {
		t.Fatalf("healthy connector restarted: %d", runner.runs)
	}
	dir, err := ManagedTunnelDir(cf, "port-test")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte("ingress:\n  - service: http_status:404\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if runner.runs != 2 {
		t.Fatalf("drift did not restart connector: %d", runner.runs)
	}
	opts.ForceRestart = true
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if runner.runs != 3 {
		t.Fatalf("forced repair did not restart: %d", runner.runs)
	}
}

// A protocol change must restart the connector exactly like an ingress change;
// otherwise the new --protocol flag never reaches the running process.
func TestAttachRestartsWhenProtocolChanges(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cf := filepath.Join(home, ".cloudflared")
	if err := os.MkdirAll(cf, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cf, id+".json"), []byte(`{"AccountTag":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}
	list, _ := json.Marshal([]TunnelInfo{{ID: id, Name: "proto-test"}})
	runner := &countingAttachRunner{list: list}
	opts := AttachOptions{
		Domain: "*.example.com", LocalURL: "unix:/tmp/relay.sock",
		TunnelName: "proto-test", ConfigDir: cf, Runner: runner, SkipDNS: true,
		Protocol: ProtocolQUIC,
	}
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if runner.runs != 1 {
		t.Fatalf("first attach did not start the connector: %d", runner.runs)
	}

	dir, err := ManagedTunnelDir(cf, "proto-test")
	if err != nil {
		t.Fatal(err)
	}
	st, err := LoadTunnelState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if st.Protocol != ProtocolQUIC {
		t.Fatalf("state protocol = %q, want quic", st.Protocol)
	}

	// Re-attaching the same protocol is a no-op: do not bounce a healthy tunnel.
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if runner.runs != 1 {
		t.Fatalf("same-protocol re-attach restarted: %d", runner.runs)
	}

	// Switching the protocol must restart.
	opts.Protocol = ProtocolHTTP2
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}
	if runner.runs != 2 {
		t.Fatalf("protocol change did not restart the connector: %d", runner.runs)
	}
}

// A caller that does not specify a protocol must not erase what the tunnel
// already recorded — otherwise the connector watchdog would reset the setting
// on every tick.
func TestAttachEmptyProtocolKeepsRecordedValue(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cf := filepath.Join(home, ".cloudflared")
	if err := os.MkdirAll(cf, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cf, id+".json"), []byte(`{"AccountTag":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}
	list, _ := json.Marshal([]TunnelInfo{{ID: id, Name: "keep-test"}})
	runner := &countingAttachRunner{list: list}
	opts := AttachOptions{
		Domain: "*.example.com", LocalURL: "unix:/tmp/relay.sock",
		TunnelName: "keep-test", ConfigDir: cf, Runner: runner, SkipDNS: true,
		Protocol: ProtocolQUIC,
	}
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}

	// Watchdog-style re-attach with no opinion about the protocol.
	opts.Protocol = ""
	if _, err := Attach(opts); err != nil {
		t.Fatal(err)
	}

	dir, err := ManagedTunnelDir(cf, "keep-test")
	if err != nil {
		t.Fatal(err)
	}
	st, err := LoadTunnelState(dir)
	if err != nil {
		t.Fatal(err)
	}
	if st.Protocol != ProtocolQUIC {
		t.Fatalf("empty Protocol overwrote the recorded value: %q", st.Protocol)
	}
	if runner.runs != 1 {
		t.Fatalf("no-op re-attach restarted the connector: %d", runner.runs)
	}
}

// An invalid protocol must be rejected before anything is written or spawned.
func TestAttachRejectsInvalidProtocol(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cf := filepath.Join(home, ".cloudflared")
	if err := os.MkdirAll(cf, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cf, id+".json"), []byte(`{"AccountTag":"x"}`), 0600); err != nil {
		t.Fatal(err)
	}
	list, _ := json.Marshal([]TunnelInfo{{ID: id, Name: "bad-proto"}})
	runner := &countingAttachRunner{list: list}
	_, err := Attach(AttachOptions{
		Domain: "*.example.com", LocalURL: "unix:/tmp/relay.sock",
		TunnelName: "bad-proto", ConfigDir: cf, Runner: runner, SkipDNS: true,
		Protocol: "tcp",
	})
	if err == nil {
		t.Fatal("Attach accepted protocol tcp")
	}
	if runner.runs != 0 {
		t.Fatalf("invalid protocol still started a connector: %d", runner.runs)
	}
}
