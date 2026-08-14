package chrome

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestExtensionDirOK(t *testing.T) {
	if ExtensionDirOK("/no/such/dir") {
		t.Fatal("expected false for missing dir")
	}
	dir := t.TempDir()
	if ExtensionDirOK(dir) {
		t.Fatal("expected false without manifest.json")
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(`{"manifest_version":3,"name":"t","version":"1"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if !ExtensionDirOK(dir) {
		t.Fatal("expected true with manifest.json")
	}
}

func TestEscapeAS(t *testing.T) {
	got := escapeAS(`a"b\c`)
	want := `a\"b\\c`
	if got != want {
		t.Fatalf("escapeAS = %q, want %q", got, want)
	}
}

func TestNormalizeOptsRequiresDir(t *testing.T) {
	_, err := normalizeOpts(LoadUnpackedOpts{})
	if err == nil {
		t.Fatal("expected error for empty ExtensionDir")
	}
}

func TestInferVersionFromDir(t *testing.T) {
	if got := InferVersionFromDir("/x/browser-agent/1.0.10"); got != "1.0.10" {
		t.Fatalf("got %q", got)
	}
	if got := InferVersionFromDir("/x/browser-agent/1.0.10/"); got != "1.0.10" {
		t.Fatalf("trailing slash got %q", got)
	}
	if got := InferVersionFromDir("/x/browser-agent/main"); got != "" {
		t.Fatalf("non-version got %q", got)
	}
}

func TestIsVersionLike(t *testing.T) {
	if !IsVersionLike("1.0.12") {
		t.Fatal("1.0.12")
	}
	if !IsVersionLike("  2.0  ") {
		t.Fatal("2.0 padded")
	}
	if IsVersionLike("Browser Agent") || IsVersionLike("v1.0.12") || IsVersionLike("") {
		t.Fatal("non-versions")
	}
}

func TestRemoveOlderEnabled(t *testing.T) {
	if !removeOlderEnabled(LoadUnpackedOpts{}) {
		t.Fatal("default should enable remove older")
	}
	if removeOlderEnabled(LoadUnpackedOpts{KeepOlder: true}) {
		t.Fatal("KeepOlder should disable")
	}
	f := false
	if removeOlderEnabled(LoadUnpackedOpts{RemoveOlder: &f}) {
		t.Fatal("RemoveOlder=false should disable")
	}
	tr := true
	if !removeOlderEnabled(LoadUnpackedOpts{RemoveOlder: &tr}) {
		t.Fatal("RemoveOlder=true should enable")
	}
	// KeepOlder wins over RemoveOlder true
	if removeOlderEnabled(LoadUnpackedOpts{KeepOlder: true, RemoveOlder: &tr}) {
		t.Fatal("KeepOlder should win")
	}
}

func TestIsTimeoutErr(t *testing.T) {
	if isTimeoutErr(nil) {
		t.Fatal("nil")
	}
	if !isTimeoutErr(context.DeadlineExceeded) {
		t.Fatal("DeadlineExceeded")
	}
	if !isTimeoutErr(fmt.Errorf("wrap: %w", context.DeadlineExceeded)) {
		t.Fatal("wrapped DeadlineExceeded")
	}
	if !isTimeoutErr(errors.New("signal: killed")) {
		t.Fatal("signal killed string")
	}
	if isTimeoutErr(errors.New("permission denied")) {
		t.Fatal("non-timeout")
	}
}
