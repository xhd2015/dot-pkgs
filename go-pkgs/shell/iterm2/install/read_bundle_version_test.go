package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadBundleVersion(t *testing.T) {
	root := t.TempDir()
	app := filepath.Join(root, "iTerm.app")
	contents := filepath.Join(app, "Contents")
	if err := os.MkdirAll(filepath.Join(contents, "MacOS"), 0o755); err != nil {
		t.Fatal(err)
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleIdentifier</key>
	<string>com.googlecode.iterm2</string>
	<key>CFBundleShortVersionString</key>
	<string>3.6.11</string>
</dict>
</plist>
`
	if err := os.WriteFile(filepath.Join(contents, "Info.plist"), []byte(plist), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := ReadBundleVersion(app)
	if err != nil {
		t.Fatal(err)
	}
	if got != "3.6.11" {
		t.Fatalf("version = %q want 3.6.11", got)
	}

	if _, err := ReadBundleVersion(""); err == nil {
		t.Fatal("empty path should error")
	}
	if _, err := ReadBundleVersion(filepath.Join(root, "missing.app")); err == nil {
		t.Fatal("missing plist should error")
	}
}
