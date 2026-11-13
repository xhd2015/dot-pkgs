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
