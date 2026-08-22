package snapshot_test

import (
	"strings"
	"testing"
	"time"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/iterm2/snapshot"
)

func TestEnrichFromProcs_BusyAndIdle(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	idle, shellPID, chosen, _, procs := snapshot.EnrichFromProcs([]snapshot.ProcRow{
		{PID: 1, PPID: 0, Stat: "Ss", Command: "login -fp u"},
		{PID: 2, PPID: 1, Stat: "S+", Command: "-zsh"},
	}, nil, now)
	if idle == nil || !*idle {
		t.Fatalf("idle=%v", idle)
	}
	if shellPID == nil || *shellPID != 2 {
		t.Fatalf("shellPID=%v", shellPID)
	}
	if chosen == nil || chosen.PID != 2 {
		t.Fatalf("chosen=%v", chosen)
	}
	if len(procs) != 2 {
		t.Fatalf("procs=%d", len(procs))
	}

	idle, _, chosen, _, _ = snapshot.EnrichFromProcs([]snapshot.ProcRow{
		{PID: 10, PPID: 0, Stat: "Ss", Command: "login -fp u"},
		{PID: 11, PPID: 10, Stat: "S", Command: "-zsh"},
		{PID: 12, PPID: 11, Stat: "R+", Command: "python train.py"},
	}, nil, now)
	if idle == nil || *idle {
		t.Fatalf("busy idle=%v", idle)
	}
	if chosen == nil || chosen.PID != 12 {
		t.Fatalf("busy chosen=%v", chosen)
	}
}

func TestListTTYProcs_Inject(t *testing.T) {
	t.Parallel()
	c := snapshot.NewCollector()
	c.ListTTYProcs = func(ttys []string) ([]snapshot.TTYProc, error) {
		if len(ttys) != 2 || ttys[0] != "ttys001" || ttys[1] != "ttys002" {
			t.Fatalf("ttys=%v", ttys)
		}
		return []snapshot.TTYProc{
			{ProcRow: snapshot.ProcRow{PID: 1, PPID: 0, Stat: "Ss", Command: "login"}, TTY: "ttys001"},
			{ProcRow: snapshot.ProcRow{PID: 2, PPID: 1, Stat: "S+", Command: "-zsh"}, TTY: "ttys001"},
			{ProcRow: snapshot.ProcRow{PID: 3, PPID: 0, Stat: "Ss", Command: "login"}, TTY: "ttys002"},
		}, nil
	}
	rows, err := c.ListTTYProcs([]string{"ttys001", "ttys002"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 {
		t.Fatalf("rows=%d", len(rows))
	}
	if rows[0].TTY != "ttys001" || rows[2].TTY != "ttys002" {
		t.Fatalf("tty tags %+v", rows)
	}
}

func TestAppTell_ListWindowsScriptUsesPath(t *testing.T) {
	t.Parallel()
	c := snapshot.NewCollector()
	c.ITermRunning = func() bool { return true }
	c.AppTell = "/Applications/iTerm.app"
	c.AppTag = "/Applications/iTerm.app"
	var saw string
	c.RunAppleScript = func(script string) (string, error) {
		saw = script
		return "###W###1###Main###42\n", nil
	}
	wins, _, err := c.ListWindows()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(saw, `tell application "/Applications/iTerm.app"`) {
		t.Fatalf("script missing path tell:\n%s", saw)
	}
	if len(wins) != 1 || wins[0].App != "/Applications/iTerm.app" {
		t.Fatalf("wins=%+v", wins)
	}
}

func TestAppTell_EmptyDefaultsToITerm2(t *testing.T) {
	t.Parallel()
	c := snapshot.NewCollector()
	c.ITermRunning = func() bool { return true }
	var saw string
	c.RunAppleScript = func(script string) (string, error) {
		saw = script
		return "", nil
	}
	_, _, err := c.ListWindows()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(saw, `tell application "iTerm2"`) {
		t.Fatalf("script missing default tell:\n%s", saw)
	}
}
