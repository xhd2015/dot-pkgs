// Package interactive starts a login-style interactive shell in a target directory.
package interactive

import (
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/bash"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/detect"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/fish"
	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/zsh"
)

// LoginInteractive starts an interactive shell in dir using detect.Shell()
// ($SHELL basename → bash|zsh|fish; default bash). promptPrefix decorates PS1
// like mvd's project name. Blocks until the shell exits.
//
// On non-zero shell exit, returns *exec.ExitError so callers can propagate the
// exit code (wrk ExitCodeError / mvd os.Exit).
func LoginInteractive(dir, promptPrefix string, extraEnv ...string) error {
	var cmd *exec.Cmd
	switch detect.Shell() {
	case "zsh":
		zdotdir, err := zsh.RcFile(promptPrefix)
		if err != nil {
			return fmt.Errorf("prepare zsh rc: %w", err)
		}
		defer os.RemoveAll(zdotdir)
		cmd = zsh.Login(dir, zdotdir, extraEnv...)
	case "fish":
		configHome, err := fish.RcFile(promptPrefix)
		if err != nil {
			return fmt.Errorf("prepare fish rc: %w", err)
		}
		defer os.RemoveAll(configHome)
		cmd = fish.Login(dir, configHome, extraEnv...)
	default:
		rcfile, err := bash.RcFile(promptPrefix)
		if err != nil {
			return fmt.Errorf("prepare bash rc: %w", err)
		}
		defer os.Remove(rcfile)
		cmd = bash.Login(dir, rcfile, extraEnv...)
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr
		}
		return fmt.Errorf("interactive shell: %w", err)
	}
	return nil
}
