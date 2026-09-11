// Package install installs a local Go CLI binary the way developers expect
// on PATH: prefer an existing LookPath hit, else ~/.local/bin; also refresh
// other existing copies under ~/.local/bin, GOBIN, and GOPATH/bin; ensure
// ~/.local/bin is on PATH via shell/localbin; ad-hoc codesign on macOS.
//
// Binary naming matches `go install`: last element of the package import
// path; for the module root package, the module path basename.
package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	shelllocalbin "github.com/xhd2015/dot-pkgs/go-pkgs/shell/localbin"
)

// Options controls a local binary install.
type Options struct {
	// Dir is the module root (working directory for go build). Required.
	Dir string
	// Package is the main package to build (e.g. "./cmd/browser-agent" or ".").
	// Required.
	Package string
	// BinName overrides the derived binary name. Empty → go-install naming.
	BinName string
	// ModulePath is the go.mod module path. Empty → read from Dir/go.mod when
	// needed to name a root-package install.
	ModulePath string

	// LookPath defaults to exec.LookPath.
	LookPath func(file string) (string, error)
	// GoEnv returns `go env <name>` (trimmed). Defaults to running go env.
	GoEnv func(name string) (string, error)
	// UserHome defaults to os.UserHomeDir.
	UserHome func() (string, error)
	// GOOS defaults to runtime.GOOS.
	GOOS string

	// Build runs `go build -o out pkg` in dir. Defaults to exec go build.
	Build func(dir, pkg, out string) error
	// CopyFile copies src to dst (overwrite). Defaults to read+write.
	CopyFile func(src, dst string) error
	// CodeSign signs path on darwin. Nil → default `codesign --force --sign -`.
	// Non-nil replaces the default entirely (including no-op on non-darwin).
	CodeSign func(path string) error
	// EnsurePATH updates shell rc so ~/.local/bin is on PATH when DestDir is
	// the default local bin. Nil → shell/localbin.EnsureOnPATH.
	EnsurePATH func(destDir string, stderr io.Writer) error

	Stdout io.Writer
	Stderr io.Writer
}

// Result describes what Install wrote.
type Result struct {
	BinName  string
	Primary  string
	Extras   []string // additional existing-copy paths refreshed (not primary)
	Signed   []string // paths that were codesigned successfully
	PATHEnsured bool  // EnsurePATH ran for default ~/.local/bin
}

// Install builds Package into the resolved primary destination, refreshes
// other existing copies (~/.local/bin, GOBIN, GOPATH/bin), ensures ~/.local/bin
// on PATH when writing there, and ad-hoc codesigns every written path on macOS.
func Install(opts Options) (Result, error) {
	var zero Result
	if strings.TrimSpace(opts.Dir) == "" {
		return zero, fmt.Errorf("install: Dir is required")
	}
	if strings.TrimSpace(opts.Package) == "" {
		return zero, fmt.Errorf("install: Package is required")
	}
	stdout := opts.Stdout
	if stdout == nil {
		stdout = io.Discard
	}
	stderr := opts.Stderr
	if stderr == nil {
		stderr = io.Discard
	}

	binName, err := resolveBinName(opts)
	if err != nil {
		return zero, err
	}

	homeFn := opts.UserHome
	if homeFn == nil {
		homeFn = os.UserHomeDir
	}
	home, herr := homeFn()
	if herr != nil {
		return zero, fmt.Errorf("install: resolve home: %w", herr)
	}
	home = strings.TrimSpace(home)
	if home == "" {
		return zero, fmt.Errorf("install: HOME is unset")
	}

	goEnv := opts.GoEnv
	if goEnv == nil {
		goEnv = defaultGoEnv
	}

	primary, fromPATH, err := ResolvePrimary(binName, home, resolveHooks{
		LookPath: opts.LookPath,
		GoEnv:    goEnv,
	})
	if err != nil {
		return zero, err
	}

	fmt.Fprintf(stdout, "==> Installing %s → %s", binName, primary)
	if fromPATH {
		fmt.Fprintf(stdout, " (existing on PATH)")
	} else {
		fmt.Fprintf(stdout, " (default ~/.local/bin)")
	}
	fmt.Fprintln(stdout)

	build := opts.Build
	if build == nil {
		build = defaultBuild
	}
	if err := os.MkdirAll(filepath.Dir(primary), 0o755); err != nil {
		return zero, fmt.Errorf("install: mkdir %s: %w", filepath.Dir(primary), err)
	}
	if err := build(opts.Dir, opts.Package, primary); err != nil {
		return zero, fmt.Errorf("install: go build -o %s: %w", primary, err)
	}

	res := Result{BinName: binName, Primary: primary}
	copyFn := opts.CopyFile
	if copyFn == nil {
		copyFn = defaultCopyFile
	}

	for _, extra := range ExtraTargets(binName, home, primary, goEnv) {
		fmt.Fprintf(stdout, "==> Also updating %s (existing copy)\n", extra)
		if err := copyFn(primary, extra); err != nil {
			return res, fmt.Errorf("install: copy to %s: %w", extra, err)
		}
		res.Extras = append(res.Extras, extra)
	}

	// Ensure ~/.local/bin on PATH when we wrote under the default local bin.
	wroteLocal := false
	localDir, _ := shelllocalbin.DefaultDir(home)
	for _, p := range writtenPaths(res) {
		if shelllocalbin.IsDefaultDest(filepath.Dir(p), home) {
			wroteLocal = true
			break
		}
	}
	if wroteLocal {
		ensure := opts.EnsurePATH
		if ensure == nil {
			ensure = defaultEnsurePATH
		}
		_ = ensure(localDir, stderr)
		res.PATHEnsured = true
	}

	goos := opts.GOOS
	if goos == "" {
		goos = runtime.GOOS
	}
	sign := opts.CodeSign
	if sign == nil {
		sign = func(path string) error {
			return defaultCodeSign(goos, path)
		}
	}
	for _, p := range writtenPaths(res) {
		if goos == "darwin" || opts.CodeSign != nil {
			fmt.Fprintf(stdout, "==> codesign %s\n", p)
			if err := sign(p); err != nil {
				fmt.Fprintf(stderr, "warning: codesign %s: %v\n", p, err)
				continue
			}
			res.Signed = append(res.Signed, p)
		}
	}

	fmt.Fprintf(stdout, "\nInstalled %s\n", res.Primary)
	for _, e := range res.Extras {
		fmt.Fprintf(stdout, "Also installed %s\n", e)
	}
	fmt.Fprintf(stdout, "Ensure %s is on your PATH", filepath.Dir(res.Primary))
	if len(res.Extras) > 0 {
		fmt.Fprint(stdout, " (earlier PATH entries may still shadow it until updated)")
	}
	fmt.Fprintln(stdout)
	return res, nil
}

func writtenPaths(res Result) []string {
	out := []string{res.Primary}
	out = append(out, res.Extras...)
	return out
}

type resolveHooks struct {
	LookPath func(file string) (string, error)
	GoEnv    func(name string) (string, error)
}

// ResolvePrimary returns where the binary should be written primarily:
// existing LookPath(bin) if found, else ~/.local/bin/<bin>.
// fromPATH is true when the dest came from LookPath.
func ResolvePrimary(binName, home string, hooks resolveHooks) (dest string, fromPATH bool, err error) {
	binName = strings.TrimSpace(binName)
	if binName == "" {
		return "", false, fmt.Errorf("install: empty binary name")
	}
	lp := hooks.LookPath
	if lp == nil {
		lp = exec.LookPath
	}
	if p, err := lp(binName); err == nil && strings.TrimSpace(p) != "" {
		abs, aerr := filepath.Abs(p)
		if aerr == nil {
			return abs, true, nil
		}
		return p, true, nil
	}
	localDir, err := shelllocalbin.DefaultDir(home)
	if err != nil {
		return "", false, fmt.Errorf("install: %w", err)
	}
	return filepath.Join(localDir, binName), false, nil
}

// ExtraTargets lists existing binary copies to refresh (not including primary).
// Candidates: ~/.local/bin/<bin>, $GOBIN/<bin>, $GOPATH/bin/<bin>.
func ExtraTargets(binName, home, primary string, goEnv func(string) (string, error)) []string {
	var cands []string
	if localDir, err := shelllocalbin.DefaultDir(home); err == nil {
		cands = append(cands, filepath.Join(localDir, binName))
	}
	if goEnv == nil {
		goEnv = defaultGoEnv
	}
	if gobin := strings.TrimSpace(mustGoEnv(goEnv, "GOBIN")); gobin != "" {
		cands = append(cands, filepath.Join(gobin, binName))
	}
	if gopath := strings.TrimSpace(mustGoEnv(goEnv, "GOPATH")); gopath != "" {
		if i := strings.IndexByte(gopath, filepath.ListSeparator); i >= 0 {
			gopath = gopath[:i]
		}
		cands = append(cands, filepath.Join(gopath, "bin", binName))
	}

	seen := map[string]bool{}
	var out []string
	for _, c := range cands {
		if sameFilePath(c, primary) {
			continue
		}
		key := filepath.Clean(c)
		if seen[key] {
			continue
		}
		if !fileExists(c) {
			continue
		}
		seen[key] = true
		out = append(out, c)
	}
	return out
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

func mustGoEnv(goEnv func(string) (string, error), name string) string {
	v, err := goEnv(name)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(v)
}

// LocalBinPath returns ~/.local/bin/<binName>.
func LocalBinPath(home, binName string) string {
	return filepath.Join(home, ".local", "bin", binName)
}

// ShouldCopyExisting reports whether primary should also be copied to path
// (path already exists as a file and is not the same as primary).
func ShouldCopyExisting(primary, path string) bool {
	if sameFilePath(primary, path) {
		return false
	}
	return fileExists(path)
}

// ShouldMirror is an alias for ShouldCopyExisting (kept for older callers).
func ShouldMirror(primary, localBin string) bool {
	return ShouldCopyExisting(primary, localBin)
}

func sameFilePath(a, b string) bool {
	a = filepath.Clean(a)
	b = filepath.Clean(b)
	if a == b {
		return true
	}
	ae, errA := filepath.EvalSymlinks(a)
	be, errB := filepath.EvalSymlinks(b)
	if errA == nil && errB == nil {
		return ae == be
	}
	return false
}

func resolveBinName(opts Options) (string, error) {
	if n := strings.TrimSpace(opts.BinName); n != "" {
		return n, nil
	}
	modPath := strings.TrimSpace(opts.ModulePath)
	if modPath == "" && needsModulePath(opts.Package) {
		p, err := readModulePath(filepath.Join(opts.Dir, "go.mod"))
		if err != nil {
			return "", fmt.Errorf("install: resolve binary name: %w", err)
		}
		modPath = p
	}
	return BinNameFromPackage(opts.Package, modPath), nil
}

func needsModulePath(pkg string) bool {
	p := strings.TrimSpace(pkg)
	return p == "" || p == "." || p == "./"
}

// BinNameFromPackage returns the go-install binary name for pkg.
// modulePath is used when pkg is the module root ("." / "./").
func BinNameFromPackage(pkg, modulePath string) string {
	p := strings.TrimSpace(pkg)
	p = strings.TrimSuffix(p, "/")
	if p == "" || p == "." || p == "./" {
		mod := strings.TrimSpace(modulePath)
		mod = strings.TrimSuffix(mod, "/")
		if mod == "" {
			return ""
		}
		return filepath.Base(mod)
	}
	p = filepath.ToSlash(p)
	p = strings.TrimPrefix(p, "./")
	for strings.HasPrefix(p, "../") {
		p = strings.TrimPrefix(p, "../")
	}
	return pathBase(p)
}

func pathBase(p string) string {
	p = strings.TrimSuffix(p, "/")
	if i := strings.LastIndex(p, "/"); i >= 0 {
		return p[i+1:]
	}
	return p
}

func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("no module line in %s", goModPath)
}

func defaultGoEnv(name string) (string, error) {
	out, err := exec.Command("go", "env", name).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func defaultBuild(dir, pkg, out string) error {
	cmd := exec.Command("go", "build", "-o", out, pkg)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultCopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, data, 0o755); err != nil {
		return err
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}

func defaultCodeSign(goos, path string) error {
	if goos != "darwin" {
		return nil
	}
	cmd := exec.Command("codesign", "--force", "--sign", "-", path)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func defaultEnsurePATH(destDir string, stderr io.Writer) error {
	shelllocalbin.EnsureOnPATH(shelllocalbin.EnsureOpts{
		DestDir: destDir,
		Stderr:  stderr,
	})
	return nil
}
