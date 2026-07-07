package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Run(ctx context.Context, dir string, args ...string) (string, error) {
	out, ok, err := RunOptional(ctx, dir, args...)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("git %s in %s returned no output", strings.Join(args, " "), dir)
	}
	return out, nil
}

func RunOptional(ctx context.Context, dir string, args ...string) (string, bool, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 && text == "" {
			return "", false, nil
		}
		return "", false, normalizeError(dir, args, err, output)
	}
	return text, true, nil
}

func RunCombined(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
	output, err := cmd.CombinedOutput()
	text := strings.TrimSpace(string(output))
	if err != nil {
		return "", normalizeError(dir, args, err, output)
	}
	return text, nil
}

func normalizeError(dir string, args []string, err error, output []byte) error {
	msg := strings.TrimSpace(string(output))
	if msg == "" {
		msg = err.Error()
	}
	msg = oneLine(msg)
	if msg == "" {
		msg = err.Error()
		msg = oneLine(msg)
	}
	return fmt.Errorf("git %s in %s: %s", strings.Join(args, " "), dir, msg)
}

func oneLine(msg string) string {
	msg = strings.TrimSpace(msg)
	if idx := strings.Index(msg, "\n"); idx >= 0 {
		msg = strings.TrimSpace(msg[:idx])
	}
	return msg
}