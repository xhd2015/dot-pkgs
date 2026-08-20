package withgo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/xhd2015/xgo/support/downloadgo"
)

// ResolveOptions controls ResolveGoroot. Install, writers, and InstallDir are
// caller-injected seams — never process-global Setenv/Chdir/stdio.
type ResolveOptions struct {
	InstallDir     string
	Download       bool
	Prompt         string
	Stdout, Stderr io.Writer
	Install        func(ctx context.Context, version, installDir string) (string, error)
}

// DefaultInstallDir is filepath.Join(home, "installed") from os.UserHomeDir.
func DefaultInstallDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "installed"), nil
}

// TargetGoroot returns the install location selected for goVersion without
// checking for it or downloading it. Callers that need to preview a command
// (for example a dry-run) can use this without causing an installation.
func TargetGoroot(goVersion string, opts ResolveOptions) string {
	pin := PinPatch(goVersion)
	return downloadgo.Target(opts.InstallDir, pin)
}

// ResolveGoroot pins goVersion and returns dest $InstallDir/<pin>.
// An existing dest directory is returned without install or Prompt.
// Missing dest with Download false is an error. Missing dest with Download
// true writes Prompt to Stderr (if set) then calls Install or downloadgo.Download.
func ResolveGoroot(goVersion string, opts ResolveOptions) (string, error) {
	pin := PinPatch(goVersion)
	dest := TargetGoroot(goVersion, opts)
	fi, err := os.Stat(dest)
	if err == nil && fi.IsDir() {
		return dest, nil
	}
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if !opts.Download {
		if err == nil {
			return "", fmt.Errorf("%s exists and is not a directory", dest)
		}
		return "", fmt.Errorf("go %s not installed at %s", pin, dest)
	}
	if opts.Prompt != "" && opts.Stderr != nil {
		fmt.Fprint(opts.Stderr, opts.Prompt)
	}
	if opts.Install != nil {
		return opts.Install(context.Background(), pin, opts.InstallDir)
	}
	return downloadgo.Download(context.Background(), pin, downloadgo.Options{
		Dir:    opts.InstallDir,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
	})
}
