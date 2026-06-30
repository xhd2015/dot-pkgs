package ptywrap

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/creack/pty"
)

func startPTY(command []string, cwd string, opts SpawnOptions) (*exec.Cmd, *os.File, []string, error) {
	if len(command) == 0 {
		return startDefaultShellPTY(cwd, opts)
	}
	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get cwd: %w", err)
		}
	}
	if info, err := os.Stat(cwd); err != nil || !info.IsDir() {
		return nil, nil, nil, fmt.Errorf("invalid working directory: %s", cwd)
	}

	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	if len(opts.ExtraPaths) > 0 {
		cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH")+":"+strings.Join(opts.ExtraPaths, ":"))
	}
	if opts.PS1 != "" {
		cmd.Env = append(cmd.Env, "PS1="+opts.PS1)
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("start pty: %w", err)
	}
	pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80})
	return cmd, ptmx, append([]string(nil), command...), nil
}

func startDefaultShellPTY(cwd string, opts SpawnOptions) (*exec.Cmd, *os.File, []string, error) {
	shellPath := "bash"
	shellFlags := []string{"--login", "-i"}
	if opts.Shell != "" {
		shellPath = opts.Shell
	}
	if len(opts.ShellFlags) > 0 {
		shellFlags = append([]string(nil), opts.ShellFlags...)
	}

	patchOpts := rcPatchOptions{
		ExtraPaths: opts.ExtraPaths,
		PS1:        opts.PS1,
	}

	var extraEnv []string
	shellBase := filepath.Base(shellPath)
	switch {
	case strings.Contains(shellBase, "zsh"):
		if zdotdir, err := writeCustomZshRC(patchOpts); err == nil {
			extraEnv = append(extraEnv, "ZDOTDIR="+zdotdir)
		}
	default:
		if rcFile, err := writeCustomBashRC(patchOpts); err == nil {
			shellFlags = replaceLoginWithRCFile(shellFlags, rcFile)
		}
	}

	if cwd == "" {
		var err error
		cwd, err = os.Getwd()
		if err != nil {
			return nil, nil, nil, fmt.Errorf("get cwd: %w", err)
		}
	}
	if info, err := os.Stat(cwd); err != nil || !info.IsDir() {
		return nil, nil, nil, fmt.Errorf("invalid working directory: %s", cwd)
	}

	cmd := exec.Command(shellPath, shellFlags...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")
	cmd.Env = append(cmd.Env, extraEnv...)
	if len(opts.ExtraPaths) > 0 {
		cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH")+":"+strings.Join(opts.ExtraPaths, ":"))
	}
	if opts.PS1 != "" {
		cmd.Env = append(cmd.Env, "PS1="+opts.PS1)
	}

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("start pty: %w", err)
	}
	pty.Setsize(ptmx, &pty.Winsize{Rows: 24, Cols: 80})
	command := append([]string{shellPath}, shellFlags...)
	return cmd, ptmx, command, nil
}