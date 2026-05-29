package zsh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RcFile creates a temporary ZDOTDIR with .zshenv and .zshrc that
// source system and user config files, then prepend prefix to PS1.
// Caller should remove the returned directory (os.RemoveAll) when done.
func RcFile(prefix string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}

	dir, err := os.MkdirTemp("", "mvd-zdotdir-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	// .zshenv is sourced first (always). Source system + user.
	zshenv := filepath.Join(dir, ".zshenv")
	realZshenv := filepath.Join(home, ".zshenv")
	if err := os.WriteFile(zshenv, []byte(fmt.Sprintf(`
if [ -f /etc/zshenv ]; then
	. /etc/zshenv
fi
if [ -f %[1]q ]; then
	. %[1]q
fi
`, realZshenv)), 0644); err != nil {
		return "", fmt.Errorf("write .zshenv: %w", err)
	}

	// .zshrc for interactive shells. Source system + user config, then
	// set PS1 and register precmd hook for dynamic prompts.
	zshrc := filepath.Join(dir, ".zshrc")
	realZprofile := filepath.Join(home, ".zprofile")
	realZshrc := filepath.Join(home, ".zshrc")
	if err := os.WriteFile(zshrc, []byte(fmt.Sprintf(`
if [ -f /etc/zprofile ]; then
	. /etc/zprofile
fi
if [ -f %[1]q ]; then
	. %[1]q
fi
if [ -f /etc/zshrc ]; then
	. /etc/zshrc
fi
if [ -f %[2]q ]; then
	. %[2]q
fi

_mvd_precmd() {
	case "${PS1-}" in
		*"(%[3]s) "*) ;;
		*) PS1="(%[3]s) ${PS1}" ;;
	esac
}
precmd_functions+=(_mvd_precmd)
`, realZprofile, realZshrc, prefix)), 0644); err != nil {
		return "", fmt.Errorf("write .zshrc: %w", err)
	}

	return dir, nil
}

// Login builds an *exec.Cmd that starts an interactive zsh shell in dir,
// using zdotdir as ZDOTDIR for custom init. extraEnv entries are
// appended to the process environment.
func Login(dir, zdotdir string, extraEnv ...string) *exec.Cmd {
	cmd := exec.Command("zsh", "-i")
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), append([]string{"ZDOTDIR=" + zdotdir}, extraEnv...)...)
	return cmd
}
