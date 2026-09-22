package cloudflare

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type scriptRunner struct {
	list []byte
	info []byte
}

func (s scriptRunner) Exec(name string, args ...string) ([]byte, error) {
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "tunnel list") {
		return s.list, nil
	}
	if strings.Contains(joined, "tunnel info") {
		return s.info, nil
	}
	if strings.Contains(joined, "tunnel create") {
		return []byte("refusing create in test"), os.ErrExist
	}
	return nil, nil
}

func TestEnsureTunnelForSessionMissingJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	list, _ := json.Marshal([]TunnelInfo{{ID: id, Name: "ai-critic-box"}})
	r := scriptRunner{
		list: list,
		info: []byte("NAME: ai-critic-box\nID: " + id + "\n"),
	}
	_, _, _, err := ensureTunnelForSession(r, "ai-critic-box")
	if err == nil || !strings.Contains(err.Error(), "credentials file not found") {
		t.Fatalf("err %v", err)
	}
}

func TestEnsureTunnelForSessionWithJSON(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	id := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	cfDir := filepath.Join(home, ".cloudflared")
	if err := os.MkdirAll(cfDir, 0o700); err != nil {
		t.Fatal(err)
	}
	cred := filepath.Join(cfDir, id+".json")
	if err := os.WriteFile(cred, []byte(`{"AccountTag":"x"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	list, _ := json.Marshal([]TunnelInfo{{ID: id, Name: "ai-critic-box"}})
	r := scriptRunner{
		list: list,
		info: []byte("NAME: ai-critic-box\nID: " + id + "\n"),
	}
	name, gotID, gotCred, err := ensureTunnelForSession(r, "ai-critic-box")
	if err != nil {
		t.Fatal(err)
	}
	if name != "ai-critic-box" || gotID != id || gotCred != cred {
		t.Fatalf("name=%s id=%s cred=%s", name, gotID, gotCred)
	}
}
