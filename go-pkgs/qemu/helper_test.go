package qemu

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHelperSrcHashStable(t *testing.T) {
	h1, err := HelperSrcHash()
	if err != nil {
		t.Fatal(err)
	}
	h2, err := HelperSrcHash()
	if err != nil {
		t.Fatal(err)
	}
	if h1 == "" || h1 != h2 {
		t.Fatalf("hash unstable: %q vs %q", h1, h2)
	}
	if len(h1) != 64 { // full sha256 hex
		t.Fatalf("want 64 hex chars, got %d (%q)", len(h1), h1)
	}
}

func TestDryRunCFUsesHelper(t *testing.T) {
	var out bytes.Buffer
	m := &Manager{
		Cfg:    WorkDevConfig(),
		Host:   fatalHost{t: t},
		DryRun: true,
		Stdout: &out,
	}
	_, err := m.CFStatus("")
	if err != nil {
		t.Fatal(err)
	}
	s := out.String()
	for _, want := range []string{
		"[dry-run] would",
		HelperPlaceholder,
		"--dir",
		"/root/qemu-guest",
		"status",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("missing %q in:\n%s", want, s)
		}
	}
	if strings.Contains(s, GuestCloudflaredScriptPlaceholder) {
		t.Fatalf("should not use shell script placeholder:\n%s", s)
	}
}

func TestEnsureHelperInstalls(t *testing.T) {
	dir := t.TempDir()
	cfg := WorkDevConfig()
	cfg.Dir = dir
	var stderr bytes.Buffer
	m := &Manager{Cfg: cfg, Host: LocalHost{}, Stderr: &stderr}
	bin, err := m.EnsureHelper()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "bin", HelperBinName)
	if bin != want {
		t.Fatalf("bin=%q want %q", bin, want)
	}
	st, err := os.Stat(bin)
	if err != nil || st.Size() < 64 {
		t.Fatalf("binary missing/small: %v %#v", err, st)
	}
	hash, err := os.ReadFile(bin + ".sha256")
	if err != nil {
		t.Fatal(err)
	}
	src, err := HelperSrcHash()
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(hash)) != src {
		t.Fatalf("stamp %q != %q", hash, src)
	}
	// Second call is a no-op (same stamp).
	stderr.Reset()
	bin2, err := m.EnsureHelper()
	if err != nil {
		t.Fatal(err)
	}
	if bin2 != bin {
		t.Fatalf("%q vs %q", bin2, bin)
	}
	if strings.Contains(stderr.String(), "installing") {
		t.Fatalf("expected skip reinstall, stderr=%s", stderr.String())
	}
}

func TestResolveHelperBinaryLinux(t *testing.T) {
	data, origin, err := ResolveHelperBinary("linux", "amd64", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) < 64 {
		t.Fatalf("too small from %s", origin)
	}
	if data[0] != 0x7f {
		t.Fatalf("want ELF magic from %s, got 0x%x", origin, data[0])
	}
}
