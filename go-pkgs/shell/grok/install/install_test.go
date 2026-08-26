package install

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestParseCheckJSON(t *testing.T) {
	raw := `{"currentVersion":"1.0.5","latestVersion":"1.0.5","updateAvailable":false,"installer":"internal","channel":"stable","autoUpdate":null,"error":null}`
	c, err := ParseCheckJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	if c.CurrentVersion != "1.0.5" || c.LatestVersion != "1.0.5" || c.UpdateAvailable || c.Channel != "stable" {
		t.Fatalf("got %+v", c)
	}

	prefixed := "Switched to alpha channel.\n" +
		`{"currentVersion":"1.0.5","latestVersion":"1.0.10","updateAvailable":true,"installer":"internal","channel":"alpha","autoUpdate":null,"error":null}`
	c, err = ParseCheckJSON(prefixed)
	if err != nil {
		t.Fatal(err)
	}
	if !c.UpdateAvailable || c.LatestVersion != "1.0.10" || c.Channel != "alpha" {
		t.Fatalf("prefixed = %+v", c)
	}

	if _, err := ParseCheckJSON(""); err == nil {
		t.Fatal("want empty error")
	}
	if _, err := ParseCheckJSON("not json"); err == nil {
		t.Fatal("want garbage error")
	}
}

func TestNeedsUpdateFromCheck(t *testing.T) {
	if NeedsUpdateFromCheck(CheckResult{CurrentVersion: "1.0.5", LatestVersion: "1.0.5"}) {
		t.Fatal("current should not need update")
	}
	if !NeedsUpdateFromCheck(CheckResult{CurrentVersion: "1.0.5", LatestVersion: "1.0.10", UpdateAvailable: true}) {
		t.Fatal("outdated should need update")
	}
	// Unparseable versions: fall back to UpdateAvailable.
	if !NeedsUpdateFromCheck(CheckResult{CurrentVersion: "dev", LatestVersion: "dev", UpdateAvailable: true}) {
		t.Fatal("want UpdateAvailable fallback")
	}
	if NeedsUpdateFromCheck(CheckResult{CurrentVersion: "dev", LatestVersion: "dev", UpdateAvailable: false}) {
		t.Fatal("want no update when flag false")
	}
}

func TestCheckUpdateInjected(t *testing.T) {
	c, err := CheckUpdate(context.Background(), CheckUpdateOpts{
		Bin: "/usr/local/bin/grok",
		RunCheck: func(ctx context.Context, bin string) (string, error) {
			if bin != "/usr/local/bin/grok" {
				t.Fatalf("bin=%q", bin)
			}
			return `{"currentVersion":"1.0.4","latestVersion":"1.0.5","updateAvailable":true,"installer":"internal","channel":"stable","autoUpdate":null,"error":null}`, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !NeedsUpdateFromCheck(c) || c.CurrentVersion != "1.0.4" {
		t.Fatalf("got %+v", c)
	}
}

func TestLocalVersionInjected(t *testing.T) {
	out, err := LocalVersion(context.Background(), LocalVersionOpts{
		Bin: "/opt/grok",
		RunVersion: func(ctx context.Context, bin string) (string, error) {
			if bin != "/opt/grok" {
				t.Fatalf("bin=%q", bin)
			}
			return "grok 1.0.5 (abc)", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	ver, err := ParseVersion(out)
	if err != nil || ver != "1.0.5" {
		t.Fatalf("ver=%q err=%v", ver, err)
	}
}

func TestUpdateCmdForBin(t *testing.T) {
	if got := UpdateCmdForBin(""); got != UpdateCmd {
		t.Fatalf("empty = %q", got)
	}
	if got := UpdateCmdForBin("/usr/local/bin/grok"); got != "/usr/local/bin/grok update" {
		t.Fatalf("abs = %q", got)
	}
	if got := UpdateCmdForBin("/tmp/my grok"); !strings.Contains(got, "update") || !strings.HasPrefix(got, "'") {
		t.Fatalf("quoted = %q", got)
	}
}

func TestUpdateFallsBackToInstall(t *testing.T) {
	var ran []string
	err := Update(context.Background(), UpdateOpts{
		Bin: "/tmp/grok",
		RunShell: func(ctx context.Context, cmd string) error {
			ran = append(ran, cmd)
			if strings.Contains(cmd, "update") {
				return fmt.Errorf("update boom")
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ran) != 2 || ran[0] != "/tmp/grok update" || ran[1] != InstallCmd {
		t.Fatalf("ran = %v", ran)
	}
}

func TestUpdateSuccessNoInstall(t *testing.T) {
	var ran []string
	err := Update(context.Background(), UpdateOpts{
		Bin: "/tmp/grok",
		RunShell: func(ctx context.Context, cmd string) error {
			ran = append(ran, cmd)
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ran) != 1 || ran[0] != "/tmp/grok update" {
		t.Fatalf("ran = %v", ran)
	}
}
