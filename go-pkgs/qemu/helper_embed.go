package qemu

import (
	"crypto/sha256"
	"encoding/hex"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/file/tmpdir"
)

//go:embed cfhelpersrc/main.go cfhelpersrc/config.go cfhelpersrc/guest.go cfhelpersrc/actions.go cfhelpersrc/go.mod.txt
var cfHelperSrcEmbed embed.FS

//go:embed prebuilt/linux_amd64/*
var prebuiltLinuxAMD64 embed.FS

//go:embed prebuilt/darwin_arm64/*
var prebuiltDarwinARM64 embed.FS

// HelperBinName is the installed basename on the qemu host.
const HelperBinName = "qemu-cf-helper"

// HelperPlaceholder is used in dry-run lines instead of dumping the binary path body.
const HelperPlaceholder = "<qemu-cf-helper>"

// HelperSrcHash returns a stable short SHA-256 of the embedded helper sources.
func HelperSrcHash() (string, error) {
	sub, err := fs.Sub(cfHelperSrcEmbed, "cfhelpersrc")
	if err != nil {
		return "", err
	}
	return contentHash(sub)
}

func contentHash(fsys fs.FS) (string, error) {
	h := sha256.New()
	var names []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		names = append(names, path)
		return nil
	})
	if err != nil {
		return "", err
	}
	sort.Strings(names)
	for _, name := range names {
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return "", err
		}
		_, _ = h.Write([]byte(name))
		_, _ = h.Write([]byte{0})
		_, _ = h.Write(data)
		_, _ = h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func helperInstallPath(cfg Config) string {
	cfg = cfg.normalized()
	return filepath.Join(cfg.Dir, "bin", HelperBinName)
}

func helperHashPath(binPath string) string {
	return binPath + ".sha256"
}

// EnsureHelper installs qemu-cf-helper under {Cfg.Dir}/bin when missing or
// source hash mismatches. Prefers embedded prebuilt for GOOS/GOARCH; falls
// back to go build from embedded sources.
func (m *Manager) EnsureHelper() (string, error) {
	cfg := m.Cfg.normalized()
	binPath := helperInstallPath(cfg)
	want, err := HelperSrcHash()
	if err != nil {
		return "", fmt.Errorf("helper source hash: %w", err)
	}
	if st, err := os.Stat(binPath); err == nil && st.Size() > 0 {
		if prev, err := os.ReadFile(helperHashPath(binPath)); err == nil && strings.TrimSpace(string(prev)) == want {
			return binPath, nil
		}
	}
	data, origin, err := m.resolveHelperBinary()
	if err != nil {
		return "", err
	}
	if m.Stderr != nil {
		fmt.Fprintf(m.Stderr, "qemu-cf-helper: installing from %s → %s\n", origin, binPath)
	}
	if err := m.writeHelperFile(binPath, data); err != nil {
		return "", err
	}
	if err := m.writeHelperFile(helperHashPath(binPath), []byte(want+"\n")); err != nil {
		return "", err
	}
	return binPath, nil
}

func (m *Manager) resolveHelperBinary() (data []byte, origin string, err error) {
	return ResolveHelperBinary(runtime.GOOS, runtime.GOARCH, m.Stderr)
}

// ResolveHelperBinary returns helper bytes for goos/goarch.
// Prefers embedded prebuilt; falls back to go build from embedded sources.
func ResolveHelperBinary(goos, goarch string, stderr io.Writer) ([]byte, string, error) {
	if data, ok := readPrebuilt(goos, goarch); ok {
		return data, "prebuilt:" + goos + "/" + goarch, nil
	}
	data, err := buildHelperFromSource(goos, goarch, stderr)
	if err != nil {
		return nil, "", err
	}
	return data, "go-build:" + goos + "/" + goarch, nil
}

func readPrebuilt(goos, goarch string) ([]byte, bool) {
	var fsys embed.FS
	var subdir string
	switch goos + "/" + goarch {
	case "linux/amd64":
		fsys = prebuiltLinuxAMD64
		subdir = "prebuilt/linux_amd64"
	case "darwin/arm64":
		fsys = prebuiltDarwinARM64
		subdir = "prebuilt/darwin_arm64"
	default:
		return nil, false
	}
	sub, err := fs.Sub(fsys, subdir)
	if err != nil {
		return nil, false
	}
	data, err := fs.ReadFile(sub, HelperBinName)
	if err != nil || len(data) < 64 {
		return nil, false
	}
	// ELF or Mach-O magic — reject placeholder text.
	if data[0] == '#' || (len(data) > 11 && string(data[:11]) == "placeholder") {
		return nil, false
	}
	if data[0] != 0x7f && data[0] != 0xcf && data[0] != 0xca && data[0] != 0xfe {
		// 0x7f ELF; Mach-O fat/thin vary — also accept if not obviously text.
		if isMostlyText(data) {
			return nil, false
		}
	}
	return data, true
}

func isMostlyText(data []byte) bool {
	n := len(data)
	if n > 256 {
		n = 256
	}
	for i := 0; i < n; i++ {
		if data[i] == 0 {
			return false
		}
	}
	return true
}

func buildHelperFromSource(goos, goarch string, stderr io.Writer) ([]byte, error) {
	sub, err := fs.Sub(cfHelperSrcEmbed, "cfhelpersrc")
	if err != nil {
		return nil, err
	}
	dir, cleanup, err := materializeHelperSrc(sub)
	if err != nil {
		return nil, err
	}
	defer cleanup()

	outPath := filepath.Join(dir, HelperBinName)
	if stderr != nil {
		fmt.Fprintf(stderr, "Building qemu-cf-helper: GOOS=%s GOARCH=%s\n", goos, goarch)
	}
	cmd := exec.Command("go", "build", "-o", outPath, ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch, "CGO_ENABLED=0")
	var out io.Writer = io.Discard
	if stderr != nil {
		out = stderr
	}
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go build qemu-cf-helper: %w", err)
	}
	return os.ReadFile(outPath)
}

func materializeHelperSrc(fsys fs.FS) (dir string, cleanup func(), err error) {
	parent := tmpdir.GetCommonTmpDir()
	dir, err = os.MkdirTemp(parent, "qemu-cf-helper-*")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { _ = os.RemoveAll(dir) }
	err = fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if path == "." {
				return nil
			}
			return os.MkdirAll(filepath.Join(dir, path), 0o755)
		}
		data, err := fs.ReadFile(fsys, path)
		if err != nil {
			return err
		}
		outName := filepath.Base(path)
		switch outName {
		case "go.mod.txt":
			outName = "go.mod"
		case "go.sum.txt":
			outName = "go.sum"
		}
		relDir := filepath.Dir(path)
		outPath := filepath.Join(dir, outName)
		if relDir != "." {
			if err := os.MkdirAll(filepath.Join(dir, relDir), 0o755); err != nil {
				return err
			}
			outPath = filepath.Join(dir, relDir, outName)
		}
		return os.WriteFile(outPath, data, 0o644)
	})
	if err != nil {
		cleanup()
		return "", nil, err
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("cfhelpersrc: go.mod.txt missing")
	}
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("cfhelpersrc: main.go missing")
	}
	return dir, cleanup, nil
}

func (m *Manager) writeHelperFile(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	mode := os.FileMode(0o644)
	if !strings.HasSuffix(path, ".sha256") {
		mode = 0o755
	}
	if err := os.WriteFile(tmp, data, mode); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// HelperFlags returns --dir/--user/--port/... flags from Config for the helper CLI.
func HelperFlags(cfg Config) []string {
	cfg = cfg.normalized()
	return []string{
		"--dir", cfg.Dir,
		"--user", cfg.User,
		"--port", fmt.Sprintf("%d", cfg.SSHPort),
		"--cf-bin", cfg.CFBin,
		"--cf-pid", cfg.CFPid,
		"--cf-log", cfg.CFLog,
		"--cert", cfg.CFCert,
		"--cf-ver", cfg.CFVersion,
		"--default-origin", cfg.CFOriginURL,
	}
}
