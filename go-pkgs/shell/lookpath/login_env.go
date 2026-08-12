package lookpath

import (
	"fmt"
	"strings"
	"time"
)

// LoginEnvOptions configures Resolve*LoginEnv(s).
// Injectables are nil for production defaults.
// Never mutates process env or cwd.
type LoginEnvOptions struct {
	Home     string
	Timeout  time.Duration
	RunLogin func(shell, command string, env []string) (stdout string, err error)
	ShellBin string // empty → "bash" or "zsh" by function
}

const envDumpCommand = "env -0"

// ResolveBashLoginEnvs returns the full environ after a bash login interactive
// shell as []string of KEY=value (os.Environ style).
func ResolveBashLoginEnvs(opts LoginEnvOptions) ([]string, error) {
	return resolveLoginEnvs("bash", opts)
}

// ResolveZshLoginEnvs returns the full environ after a zsh login interactive
// shell as []string of KEY=value (os.Environ style).
func ResolveZshLoginEnvs(opts LoginEnvOptions) ([]string, error) {
	return resolveLoginEnvs("zsh", opts)
}

// ResolveBashLoginEnv returns a single variable from a bash login shell dump.
// Empty name → error. Unset or empty value → ("", nil). Run failure → error.
func ResolveBashLoginEnv(name string, opts LoginEnvOptions) (string, error) {
	return resolveLoginEnv(name, "bash", opts)
}

// ResolveZshLoginEnv returns a single variable from a zsh login shell dump.
// Empty name → error. Unset or empty value → ("", nil). Run failure → error.
func ResolveZshLoginEnv(name string, opts LoginEnvOptions) (string, error) {
	return resolveLoginEnv(name, "zsh", opts)
}

func resolveLoginEnvs(defaultShell string, opts LoginEnvOptions) ([]string, error) {
	stdout, err := runLoginEnvDump(defaultShell, opts)
	if err != nil {
		return nil, err
	}
	return parseEnv0(stdout), nil
}

func resolveLoginEnv(name, defaultShell string, opts LoginEnvOptions) (string, error) {
	if name == "" {
		return "", fmt.Errorf("lookpath: empty env name")
	}
	envs, err := resolveLoginEnvs(defaultShell, opts)
	if err != nil {
		return "", err
	}
	// Missing or empty value → ("", nil) so callers can cascade.
	return lookupEnvValue(envs, name), nil
}

func runLoginEnvDump(defaultShell string, opts LoginEnvOptions) (string, error) {
	shell := opts.ShellBin
	if shell == "" {
		shell = defaultShell
	}

	runLogin := opts.RunLogin
	if runLogin == nil {
		timeout := opts.Timeout
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		runLogin = defaultRunLogin(timeout)
	}

	env := minimalLoginEnv(opts.Home)
	return runLogin(shell, envDumpCommand, env)
}

// parseEnv0 parses env -0 style stdout: NUL-delimited KEY=value records.
// Empty segments (including a trailing NUL) are skipped.
func parseEnv0(stdout string) []string {
	if stdout == "" {
		return nil
	}
	parts := strings.Split(stdout, "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	return out
}

// lookupEnvValue returns the value for name in KEY=value entries.
// Missing or empty value yields "".
func lookupEnvValue(envs []string, name string) string {
	for _, e := range envs {
		key, val, ok := strings.Cut(e, "=")
		if !ok {
			continue
		}
		if key == name {
			return val
		}
	}
	return ""
}
