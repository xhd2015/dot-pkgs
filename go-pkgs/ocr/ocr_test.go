//go:build darwin

package ocr

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecognizeFileMissing(t *testing.T) {
	_, err := RecognizeFile(context.Background(), filepath.Join(t.TempDir(), "nope.png"), Options{})
	if err == nil || !strings.Contains(err.Error(), "no such file") {
		t.Fatalf("err = %v", err)
	}
}

func TestRecognizeFileWithHelperPath(t *testing.T) {
	img := filepath.Join(t.TempDir(), "in.png")
	if err := os.WriteFile(img, []byte("not-a-real-png"), 0644); err != nil {
		t.Fatal(err)
	}
	helper := writeFakeHelper(t, "hello from helper\n")
	res, err := RecognizeFile(context.Background(), img, Options{HelperPath: helper})
	if err != nil {
		t.Fatal(err)
	}
	if res.Engine != EngineVision {
		t.Fatalf("engine = %q", res.Engine)
	}
	if res.Text != "hello from helper" {
		t.Fatalf("text = %q", res.Text)
	}
}

func TestRecognizeBytesWithHelperPath(t *testing.T) {
	helper := writeFakeHelper(t, "bytes ok\n")
	res, err := RecognizeBytes(context.Background(), []byte("png-bytes"), "png", Options{HelperPath: helper})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "bytes ok" {
		t.Fatalf("text = %q", res.Text)
	}
}

func TestEnsureVisionHelperCompileOnce(t *testing.T) {
	cache := t.TempDir()
	swiftc, counterPath := writeFakeSwiftc(t)

	opts := Options{
		CacheDir: cache,
		Swiftc:   swiftc,
	}
	ctx := context.Background()
	p1, err := ensureVisionHelper(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	p2, err := ensureVisionHelper(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	if p1 != p2 {
		t.Fatalf("paths differ: %q vs %q", p1, p2)
	}
	b, err := os.ReadFile(counterPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(string(b)); got != "1" {
		t.Fatalf("swiftc invocations = %q, want 1", got)
	}
	st, err := os.Stat(p1)
	if err != nil {
		t.Fatal(err)
	}
	if st.Mode()&0111 == 0 {
		t.Fatalf("helper not executable: %v", st.Mode())
	}
}

func writeFakeHelper(t *testing.T, stdout string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fake-helper")
	script := "#!/bin/sh\nprintf '%s' '" + strings.ReplaceAll(stdout, "'", `'\''`) + "'\n"
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeFakeSwiftc writes a swiftc stand-in that copies a tiny executable to -o.
// Returns (swiftcPath, counterPath).
func writeFakeSwiftc(t *testing.T) (string, string) {
	t.Helper()
	dir := t.TempDir()
	counter := filepath.Join(dir, "count")
	swiftc := filepath.Join(dir, "swiftc")
	helperBody := "#!/bin/sh\necho fake-vision\n"
	helperSrc := filepath.Join(dir, "payload.sh")
	if err := os.WriteFile(helperSrc, []byte(helperBody), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(counter, []byte("0"), 0644); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
set -e
count_file="` + counter + `"
n=$(cat "$count_file")
n=$((n+1))
echo "$n" > "$count_file"
out=""
while [ $# -gt 0 ]; do
  if [ "$1" = "-o" ]; then
    shift
    out="$1"
  fi
  shift
done
if [ -z "$out" ]; then
  echo "fake swiftc: missing -o" >&2
  exit 1
fi
cp "` + helperSrc + `" "$out"
chmod +x "$out"
`
	if err := os.WriteFile(swiftc, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	if err := exec.Command(swiftc, "-O", "-o", filepath.Join(dir, "x"), "y.swift").Run(); err != nil {
		t.Fatalf("fake swiftc smoke: %v", err)
	}
	_ = os.Remove(filepath.Join(dir, "x"))
	if err := os.WriteFile(counter, []byte("0"), 0644); err != nil {
		t.Fatal(err)
	}
	return swiftc, counter
}
