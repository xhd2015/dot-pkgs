package fish

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// RcFile creates a temporary XDG_CONFIG_HOME directory containing
// fish/config.fish that sources system and user config, then prepends
// prefix to the prompt by wrapping fish_prompt.
// Caller should remove the returned directory (os.RemoveAll) when done.
func RcFile(prefix string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}

	dir, err := os.MkdirTemp("", "mvd-fish-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}

	fishDir := filepath.Join(dir, "fish")
	if err := os.Mkdir(fishDir, 0755); err != nil {
		return "", fmt.Errorf("create fish dir: %w", err)
	}

	realConfig := filepath.Join(home, ".config", "fish", "config.fish")
	realFunctions := filepath.Join(home, ".config", "fish", "functions")
	configFile := filepath.Join(fishDir, "config.fish")

	content := fmt.Sprintf(`# Source system config
if test -f /etc/fish/config.fish
    source /etc/fish/config.fish
end

# Prepend real functions dir
if test -d %[1]q
    set -p fish_function_path %[1]q
end

# Source real user config
if test -f %[2]q
    source %[2]q
end

# Wrap fish_prompt to prepend prefix
if functions -q fish_prompt
    functions -c fish_prompt __mvd_old_fish_prompt
end

function fish_prompt
    echo -n "(%[3]s) "
    if functions -q __mvd_old_fish_prompt
        __mvd_old_fish_prompt
    else
        echo -n (prompt_pwd) '> '
    end
end
`, realFunctions, realConfig, prefix)

	if err := os.WriteFile(configFile, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("write config.fish: %w", err)
	}

	return dir, nil
}

// Login builds an *exec.Cmd that starts an interactive fish shell in dir,
// using configHome as XDG_CONFIG_HOME for custom init. extraEnv entries
// are appended to the process environment.
func Login(dir, configHome string, extraEnv ...string) *exec.Cmd {
	cmd := exec.Command("fish", "-i")
	cmd.Dir = dir
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), append([]string{"XDG_CONFIG_HOME=" + configHome}, extraEnv...)...)
	return cmd
}
