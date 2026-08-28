package github

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type ghUserWire struct {
	Login string `json:"login"`
}

// EnsureAuthenticated verifies gh is logged in and returns the user login.
func EnsureAuthenticated(ctx context.Context, ghBin string) (string, error) {
	bin, err := resolveGhBin(ghBin)
	if err != nil {
		return "", err
	}

	stdout, stderr, err := outputWithETXTBSYRetry(ctx, bin, "api", "user")
	if err != nil {
		return "", wrapGhAuthError(err, stderr)
	}

	var user ghUserWire
	if err := json.Unmarshal(stdout, &user); err != nil {
		return "", fmt.Errorf("decode gh api user JSON: %w", err)
	}
	if strings.TrimSpace(user.Login) == "" {
		return "", fmt.Errorf("gh api user: missing login in response")
	}
	return user.Login, nil
}

func wrapGhAuthError(err error, stderr string) error {
	stderr = strings.TrimSpace(stderr)

	if exitErr, ok := err.(*exec.ExitError); ok {
		code := exitErr.ExitCode()
		lower := strings.ToLower(stderr)
		if code == 4 && strings.Contains(lower, "auth") {
			if stderr != "" {
				return fmt.Errorf("%s: run `gh auth login` to authenticate", stderr)
			}
			return fmt.Errorf("gh auth login required")
		}
		if stderr != "" {
			return fmt.Errorf("gh api user: %s", stderr)
		}
		return fmt.Errorf("gh api user: %w", err)
	}

	if pathErr, ok := err.(*exec.Error); ok {
		if pathErr.Err == exec.ErrNotFound {
			return fmt.Errorf("gh not found")
		}
		if strings.Contains(pathErr.Error(), "no such file") || strings.Contains(pathErr.Error(), "not found") {
			return fmt.Errorf("gh not found: %w", err)
		}
	}

	if stderr != "" {
		return fmt.Errorf("gh api user: %s: %w", stderr, err)
	}
	return fmt.Errorf("gh api user: %w", err)
}