package withgo

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// ExecOptions controls Exec. Env and Dir are set on the child only.
type ExecOptions struct {
	Dir            string
	ExtraEnv       []string
	Stdin          io.Reader
	Stdout, Stderr io.Writer
}

// Exec runs args under GOROOT=abs(goroot) and PATH=$abs/bin:$PATH from
// os.Getenv (no process Setenv). ExtraEnv is appended. Bare "go" becomes
// $GOROOT/bin/go when that file exists. Empty args run env.
func Exec(goroot string, args []string, opts ExecOptions) error {
	abs, err := filepath.Abs(goroot)
	if err != nil {
		return err
	}

	cmdName, cmdArgs := execArgs(abs, args)
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Dir = opts.Dir
	cmd.Stdin = opts.Stdin
	cmd.Stdout = opts.Stdout
	cmd.Stderr = opts.Stderr

	path := filepath.Join(abs, "bin") + string(os.PathListSeparator) + os.Getenv("PATH")
	env := append(os.Environ(), "GOROOT="+abs, "PATH="+path)
	if len(opts.ExtraEnv) > 0 {
		env = append(env, opts.ExtraEnv...)
	}
	cmd.Env = env
	return cmd.Run()
}

func execArgs(absGoroot string, args []string) (string, []string) {
	if len(args) == 0 {
		return "env", nil
	}
	name := args[0]
	rest := args[1:]
	if name == "go" {
		goBin := filepath.Join(absGoroot, "bin", "go")
		if fi, err := os.Stat(goBin); err == nil && !fi.IsDir() {
			return goBin, rest
		}
	}
	return name, rest
}

// Run resolves goVersion then Execs args under that GOROOT.
func Run(goVersion string, args []string, resolve ResolveOptions, execOpts ExecOptions) error {
	goroot, err := ResolveGoroot(goVersion, resolve)
	if err != nil {
		return err
	}
	return Exec(goroot, args, execOpts)
}
