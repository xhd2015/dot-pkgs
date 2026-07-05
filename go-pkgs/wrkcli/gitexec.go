package wrkcli

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

var invocationVerbose bool

func setInvocationVerbose(v bool) {
	invocationVerbose = v
}

func logGitCommand(args []string) {
	if !invocationVerbose || !isMajorGitCommand(args) {
		return
	}
	ts := time.Now().Format("2006-01-02 15:04:05")
	fmt.Fprintf(os.Stderr, "[%s] $ git %s\n", ts, strings.Join(args, " "))
}

func isMajorGitCommand(args []string) bool {
	i := 0
	for i < len(args) && args[i] == "-C" {
		i += 2
	}
	if i >= len(args) {
		return false
	}
	switch args[i] {
	case "fetch", "checkout", "merge", "rebase", "stash":
		return true
	case "worktree":
		if i+1 < len(args) {
			switch args[i+1] {
			case "add", "remove", "move":
				return true
			}
		}
	case "branch":
		if i+1 < len(args) {
			switch args[i+1] {
			case "-D", "-m", "-b":
				return true
			}
		}
	}
	return false
}

func gitCommand(args ...string) *exec.Cmd {
	logGitCommand(args)
	return exec.Command("git", args...)
}

func gitCommandDir(repoPath string, args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	return cmd
}

func gitCommandWithEnv(repoPath string, extraEnv []string, args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", repoPath}, args...)
	logGitCommand(fullArgs)
	cmd := exec.Command("git", append([]string{"-C", repoPath}, args...)...)
	cmd.Env = append(os.Environ(), extraEnv...)
	return cmd
}