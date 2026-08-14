package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	xgocmd "github.com/xhd2015/xgo/support/cmd"
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
	// Combined stdout+stderr capture preserves prior CombinedOutput error shapes.
	var buf bytes.Buffer
	err := xgocmd.Dir(dir).
		Env([]string{"GIT_OPTIONAL_LOCKS=0"}).
		Stdout(&buf).
		Stderr(&buf).
		Run("git", args...)
	// Note: context is currently unused by xgo/support/cmd; callers still pass it
	// for API compatibility. Cancellation would require os/exec.CommandContext.
	_ = ctx
	text := strings.TrimSpace(buf.String())
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 1 && text == "" {
			return "", false, nil
		}
		return "", false, normalizeError(dir, args, err, buf.Bytes())
	}
	return text, true, nil
}

func RunCombined(ctx context.Context, dir string, args ...string) (string, error) {
	return RunEnv(ctx, dir, nil, args...)
}

// RunEnv is like RunCombined but appends extra env vars on this invocation only.
func RunEnv(ctx context.Context, dir string, extraEnv []string, args ...string) (string, error) {
	var buf bytes.Buffer
	env := append([]string{"GIT_OPTIONAL_LOCKS=0"}, extraEnv...)
	err := xgocmd.Dir(dir).
		Env(env).
		Stdout(&buf).
		Stderr(&buf).
		Run("git", args...)
	_ = ctx
	text := strings.TrimSpace(buf.String())
	if err != nil {
		return "", normalizeError(dir, args, err, buf.Bytes())
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
