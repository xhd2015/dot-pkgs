package bash

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRcFilePrefixesPS1(t *testing.T) {
	tmpHome := t.TempDir()
	bashrc := filepath.Join(tmpHome, ".bashrc")
	if err := os.WriteFile(bashrc, []byte(`PS1='TEST_PS1_VALUE'`+"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	rcfile, err := RcFile("mytest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(rcfile)

	cmd := exec.Command("bash", "--rcfile", rcfile, "-i")
	cmd.Stdin = strings.NewReader(`[[ $PS1 == "(mytest) "* ]]; exit
`)
	cmd.Env = append(os.Environ(), "HOME="+tmpHome)

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("PS1 missing prefix, exit code %d", exitErr.ExitCode())
		}
		t.Fatalf("bash failed: %v", err)
	}
}

// TestRcFileWrapsPROMPT_COMMAND verifies the prefix survives when the
// user has a PROMPT_COMMAND that dynamically rebuilds PS1 (e.g. starship,
// powerlevel10k, git-prompt).  We source /etc/profile first (which may
// inject VS Code shell integration), then user bashrc which sets a known
// PROMPT_COMMAND, then verify PS1 still contains our prefix.
func TestRcFileWrapsPROMPT_COMMAND(t *testing.T) {
	tmpHome := t.TempDir()

	bashrc := filepath.Join(tmpHome, ".bashrc")
	if err := os.WriteFile(bashrc, []byte(`
__dyn_ps1() { PS1="DYNAMIC_PS1"; }
PROMPT_COMMAND=__dyn_ps1
`), 0644); err != nil {
		t.Fatal(err)
	}

	rcfile, err := RcFile("mytest")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(rcfile)

	cmd := exec.Command("bash", "--rcfile", rcfile, "-i")
	// Print PS1 after the first PROMPT_COMMAND cycle ran.
	cmd.Stdin = strings.NewReader(`printf '%s' "$PS1"; exit
`)
	cmd.Env = append(os.Environ(), "HOME="+tmpHome)

	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			t.Fatalf("bash exited %d, output: %q", exitErr.ExitCode(), string(out))
		}
		t.Fatalf("bash failed: %v", err)
	}

	got := string(out)
	if !strings.Contains(got, "(mytest)") {
		t.Errorf("PS1 should contain prefix.\n  got: %q", got)
	}
}
