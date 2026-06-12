package main

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

func cmdCd(src string) error {
	hist, aliases, err := loadHistory()
	if err != nil {
		return err
	}

	_, _, lastLoc, err := resolveMoveSource(hist, aliases, src)
	if err != nil {
		return err
	}

	return launchShell(lastLoc, src)
}

func launchShell(dir string, name string) error {
	extraEnv := []string{"MVD_SHELL=1", "MVD_PROJECT=" + name}

	var cmd *exec.Cmd
	switch detect.Shell() {
	case "zsh":
		zdotdir, err := zsh.RcFile(name)
		if err != nil {
			return fmt.Errorf("prepare zsh rc: %w", err)
		}
		defer os.RemoveAll(zdotdir)
		cmd = zsh.Login(dir, zdotdir, extraEnv...)
	case "fish":
		configHome, err := fish.RcFile(name)
		if err != nil {
			return fmt.Errorf("prepare fish rc: %w", err)
		}
		defer os.RemoveAll(configHome)
		cmd = fish.Login(dir, configHome, extraEnv...)
	default:
		rcfile, err := bash.RcFile(name)
		if err != nil {
			return fmt.Errorf("prepare bash rc: %w", err)
		}
		defer os.Remove(rcfile)
		cmd = bash.Login(dir, rcfile, extraEnv...)
	}

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("cd: %w", err)
	}
	return nil
}
