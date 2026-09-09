package qemu

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseKV(t *testing.T) {
	in := "10.91.186.143:15000: qemu_alive=yes\n\x1b[32m10.91.186.143:15000: qemu_pid=331454\n"
	kv := ParseKV(in)
	if kv["qemu_alive"] != "yes" || kv["qemu_pid"] != "331454" {
		t.Fatalf("%v", kv)
	}
}

func TestFileConfigMissingDisabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing-qemu.json")
	if IsEnabled(path) {
		t.Fatal("missing file should be disabled")
	}
	if _, err := LoadFileConfig(path); err == nil {
		t.Fatal("expected load error for missing file")
	}
}

func TestFileConfigEnabled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "qemu.json")
	if err := SaveFileConfig(path, FileConfig{Enabled: true}); err != nil {
		t.Fatal(err)
	}
	cfg, err := LoadFileConfig(path)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Enabled {
		t.Fatalf("%+v", cfg)
	}
	if !IsEnabled(path) {
		t.Fatal("expected enabled")
	}
}

func TestDefaultFilePath(t *testing.T) {
	got := DefaultFilePath("/home/x")
	want := filepath.Join("/home/x", ".ai-critic", "qemu.json")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

type fatalHost struct {
	t *testing.T
}

func (h fatalHost) Run(name string, args ...string) error {
	h.t.Fatalf("Host.Run called: %s %v", name, args)
	return nil
}

func (h fatalHost) Output(name string, args ...string) (string, error) {
	h.t.Fatalf("Host.Output called: %s %v", name, args)
	return "", nil
}

func TestDryRunStart(t *testing.T) {
	var out bytes.Buffer
	m := &Manager{
		Cfg:    WorkDevConfig(),
		Host:   fatalHost{t: t},
		DryRun: true,
		Stdout: &out,
	}
	kv, err := m.Start()
	if err != nil {
		t.Fatal(err)
	}
	if len(kv) != 0 {
		t.Fatalf("dry-run kv=%v", kv)
	}
	s := out.String()
	for _, want := range []string{
		"[dry-run] would sh -c",
		GuestStartScriptPlaceholder,
		"-- start",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
}

func TestAiCriticConfigDir(t *testing.T) {
	cfg := AiCriticConfig()
	if !strings.HasSuffix(cfg.Dir, filepath.Join(".ai-critic", "qemu")) {
		t.Fatalf("Dir=%q", cfg.Dir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(home, ".ai-critic", "qemu")
	if cfg.Dir != want {
		t.Fatalf("Dir=%q want %q", cfg.Dir, want)
	}
	if cfg.Accel != "auto" {
		t.Fatalf("Accel=%q", cfg.Accel)
	}
	if cfg.CFOriginURL != "http://10.0.2.2:23712" {
		t.Fatalf("CFOriginURL=%q", cfg.CFOriginURL)
	}
}

func TestGuestStartScriptContainsDir(t *testing.T) {
	cfg := WorkDevConfig()
	cfg.Dir = "/tmp/custom-qemu-guest"
	s := cfg.GuestStartScript()
	if !strings.Contains(s, "DIR=/tmp/custom-qemu-guest") {
		t.Fatalf("script missing configured Dir:\n%s", s[:min(400, len(s))])
	}
	if strings.Contains(s, `DIR="${QEMU_GUEST_DIR:`) {
		t.Fatal("expected concrete DIR assignment, still has env form")
	}
}

func TestHumanBytes(t *testing.T) {
	if HumanBytes("") != "missing" {
		t.Fatal(HumanBytes(""))
	}
	if HumanBytes("69") != "69B" {
		t.Fatal(HumanBytes("69"))
	}
	if got := HumanBytes("72351744"); got != "69MiB" {
		t.Fatalf("got %s", got)
	}
}

func TestGuestSSHArgs(t *testing.T) {
	cfg := WorkDevConfig()
	args := GuestSSHArgs(cfg, false)
	joined := strings.Join(args, " ")
	for _, want := range []string{
		"ssh",
		"/root/qemu-guest/guest_ed25519",
		"-p", "22221",
		"BatchMode=yes",
		"debian@127.0.0.1",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in %v", want, args)
		}
	}
	tty := GuestSSHArgs(cfg, true)
	if !strings.Contains(strings.Join(tty, " "), "-tt") {
		t.Fatalf("%v", tty)
	}
}
