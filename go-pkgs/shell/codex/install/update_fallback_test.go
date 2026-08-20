package install

import (
	"context"
	"fmt"
	"testing"
)

func TestUpdateNoBinSuccess(t *testing.T) {
	t.Parallel()
	var calls []string
	err := Update(context.Background(), UpdateOpts{
		RunShell: func(ctx context.Context, cmd string) error {
			_ = ctx
			calls = append(calls, cmd)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertStringSlice(t, "ShellCalls", calls, []string{UpdateCmd})
}

func TestUpdateWithBinPathQualifies(t *testing.T) {
	t.Parallel()
	var calls []string
	err := Update(context.Background(), UpdateOpts{
		Bin: "/opt/homebrew/bin/codex",
		RunShell: func(ctx context.Context, cmd string) error {
			_ = ctx
			calls = append(calls, cmd)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertStringSlice(t, "ShellCalls", calls, []string{"/opt/homebrew/bin/codex update"})
}

func TestUpdateFailThenInstallOK(t *testing.T) {
	t.Parallel()
	var calls []string
	err := Update(context.Background(), UpdateOpts{
		Bin: "/tmp/codex",
		RunShell: func(ctx context.Context, cmd string) error {
			_ = ctx
			calls = append(calls, cmd)
			if len(calls) == 1 {
				return fmt.Errorf("injected update failure")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assertStringSlice(t, "ShellCalls", calls, []string{"/tmp/codex update", InstallCmd})
}

func TestUpdateThenInstallBothFail(t *testing.T) {
	t.Parallel()
	var calls []string
	err := Update(context.Background(), UpdateOpts{
		RunShell: func(ctx context.Context, cmd string) error {
			_ = ctx
			calls = append(calls, cmd)
			return fmt.Errorf("injected fail: %s", cmd)
		},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	assertStringSlice(t, "ShellCalls", calls, []string{UpdateCmd, InstallCmd})
}

func TestEnsureOutdatedUpdateFailInstalls(t *testing.T) {
	t.Parallel()
	bin := "/tmp/ensure-codex"
	var calls []string
	res, err := Ensure(context.Background(), EnsureOpts{
		LookPath: func(file string) (string, error) {
			_ = file
			return bin, nil
		},
		RunVersion: func(ctx context.Context, b string) (string, error) {
			_ = ctx
			_ = b
			return "codex-cli 0.1.0", nil
		},
		FetchLatest: func(ctx context.Context) (string, error) {
			_ = ctx
			return "0.2.0", nil
		},
		RunShell: func(ctx context.Context, cmd string) error {
			_ = ctx
			calls = append(calls, cmd)
			if len(calls) == 1 {
				return fmt.Errorf("injected update failure")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "update" {
		t.Fatalf("Action = %q, want update", res.Action)
	}
	assertStringSlice(t, "ShellCalls", calls, []string{bin + " update", InstallCmd})
}
