package bash

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RcFile creates a temporary bash init file that sources the system
// profile, user's ~/.bash_profile and ~/.bashrc (if they exist), then
// prepends prefix to PS1 — similar to Python venv's activate.
//
// The caller is responsible for removing the returned file when done.
func RcFile(prefix string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}

	f, err := os.CreateTemp("", "mvd-rc-*.sh")
	if err != nil {
		return "", fmt.Errorf("create temp: %w", err)
	}
	defer f.Close()

	// Mimic login shell: system profile first, then user files.
	fmt.Fprintf(f, `
if [ -f /etc/profile ]; then
	. /etc/profile
fi
`)
	for _, name := range []string{".bash_profile", ".bashrc"} {
		p := filepath.Join(home, name)
		fmt.Fprintf(f, `
if [ -f %q ]; then
	. %q
fi
`, p, p)
	}

	fmt.Fprintf(f, `
_mvd_prefix_ps1() {
	case "${PS1-}" in
		*"(%[1]s) "*) ;;
		*) PS1="(%[1]s) ${PS1}" ;;
	esac
}

__mvd_orig_prompt() {
	eval "$_MVD_OLD_PROMPT_COMMAND"
}

if [ -n "${PROMPT_COMMAND-}" ]; then
	_MVD_OLD_PROMPT_COMMAND="$PROMPT_COMMAND"
	PROMPT_COMMAND='__mvd_orig_prompt; _mvd_prefix_ps1'
else
	PROMPT_COMMAND=_mvd_prefix_ps1
fi
`, prefix)

	return f.Name(), nil
}

// Login builds an *exec.Cmd that starts an interactive bash shell in dir,
// sourcing rcfile for init. extraEnv entries are appended to the process
// environment.
func Login(dir, rcfile string, extraEnv ...string) *exec.Cmd {
	cmd := exec.Command("bash", "--rcfile", rcfile, "-i")
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd
}
