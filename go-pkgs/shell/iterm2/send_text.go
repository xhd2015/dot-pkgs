package iterm2

import (
	"fmt"
	"strings"

	"github.com/xhd2015/dot-pkgs/go-pkgs/shell/applescript"
)

// SendTextOptions controls how text is written into a session.
type SendTextOptions struct {
	// Focus, when true, activates iTerm2 and selects the window/tab/session
	// before writing. Default false: write without switching focus.
	Focus bool
	// NoSubmit appends AppleScript "without newline" (stage; no Enter).
	NoSubmit bool
	// NoCtrlU disables the default Ctrl-U (ASCII 21) prefix that clears
	// leftover keystrokes before the text.
	NoCtrlU bool
}

// SendTextConfig injects Exec / app discovery for tests.
type SendTextConfig struct {
	// Exec runs AppleScript. Nil uses a stderr-capturing osascript runner.
	Exec func(script string) error
	// Apps, when non-nil, is the exact tell list (already running). Empty means none.
	Apps   []ContentsApp
	Getenv func(key string) string
	Home   func() string
	IsApp  func(path string) bool
	// Running reports whether abs bundle is a live process. Nil uses pgrep.
	Running func(abs string) bool
}

// BuildSendTextScript returns AppleScript that locates sessionID and writes text.
// When opts.Focus is false, it does not activate or select. appPath is a
// filesystem path for TellApplicationHeader (empty → bare "iTerm2").
//
// Iteration uses count snapshots and per-item try/on error so mid-scan
// window/tab/session churn (AppleScript -1719 Invalid index) skips the
// missing item instead of aborting the whole write.
func BuildSendTextScript(sessionID, text string, opts SendTextOptions, appPath string) string {
	uuid := SessionUUID(sessionID)
	escapedUUID := EscapePathForAppleScript(uuid)
	writeLine := buildSendWriteTextLine(text, opts)

	if opts.Focus {
		return buildSendTextFocusScript(escapedUUID, writeLine, appPath)
	}
	return buildSendTextNoFocusScript(escapedUUID, writeLine, appPath)
}

func buildSendWriteTextLine(text string, opts SendTextOptions) string {
	body := applescript.WriteTextExpr(text)
	var expr string
	if opts.NoCtrlU {
		expr = body
	} else {
		expr = `((ASCII character 21) & ` + body + `)`
	}
	line := `write text ` + expr
	if opts.NoSubmit {
		line += ` without newline`
	}
	return line
}

func buildSendTextNoFocusScript(escapedUUID, writeLine, appPath string) string {
	lines := []string{
		TellApplicationHeader(appPath),
		`  set windowCount to 0`,
		`  try`,
		`    set windowCount to count of windows`,
		`  on error`,
		`  end try`,
		`  repeat with wi from 1 to windowCount`,
		`    try`,
		`      set aWindow to window wi`,
		`      set tabCount to 0`,
		`      try`,
		`        set tabCount to count of tabs of aWindow`,
		`      on error`,
		`      end try`,
		`      repeat with ti from 1 to tabCount`,
		`        try`,
		`          set aTab to tab ti of aWindow`,
		`          set sessionCount to 0`,
		`          try`,
		`            set sessionCount to count of sessions of aTab`,
		`          on error`,
		`          end try`,
		`          repeat with si from 1 to sessionCount`,
		`            try`,
		`              set aSession to session si of aTab`,
		`              set sid to ""`,
		`              try`,
		`                set sid to id of aSession as string`,
		`              on error`,
		`              end try`,
		`              if sid contains "` + escapedUUID + `" then`,
		`                tell aSession`,
		`                  ` + writeLine,
		`                end tell`,
		`                return`,
		`              end if`,
		`            on error`,
		`            end try`,
		`          end repeat`,
		`        on error`,
		`        end try`,
		`      end repeat`,
		`    on error`,
		`    end try`,
		`  end repeat`,
		`  error "session not found: ` + escapedUUID + `"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

func buildSendTextFocusScript(escapedUUID, writeLine, appPath string) string {
	lines := []string{
		TellApplicationHeader(appPath),
		`  activate`,
		`  set targetSession to missing value`,
		`  set targetTab to missing value`,
		`  set targetWindow to missing value`,
		`  set windowCount to 0`,
		`  try`,
		`    set windowCount to count of windows`,
		`  on error`,
		`  end try`,
		`  repeat with wi from 1 to windowCount`,
		`    try`,
		`      set aWindow to window wi`,
		`      set tabCount to 0`,
		`      try`,
		`        set tabCount to count of tabs of aWindow`,
		`      on error`,
		`      end try`,
		`      repeat with ti from 1 to tabCount`,
		`        try`,
		`          set aTab to tab ti of aWindow`,
		`          set sessionCount to 0`,
		`          try`,
		`            set sessionCount to count of sessions of aTab`,
		`          on error`,
		`          end try`,
		`          repeat with si from 1 to sessionCount`,
		`            try`,
		`              set aSession to session si of aTab`,
		`              set sid to ""`,
		`              try`,
		`                set sid to id of aSession as string`,
		`              on error`,
		`              end try`,
		`              if sid contains "` + escapedUUID + `" then`,
		`                set targetSession to aSession`,
		`                set targetTab to aTab`,
		`                set targetWindow to aWindow`,
		`                exit repeat`,
		`              end if`,
		`            on error`,
		`            end try`,
		`          end repeat`,
		`          if targetSession is not missing value then exit repeat`,
		`        on error`,
		`        end try`,
		`      end repeat`,
		`      if targetSession is not missing value then exit repeat`,
		`    on error`,
		`    end try`,
		`  end repeat`,
		`  if targetSession is missing value then`,
		`    error "session not found: ` + escapedUUID + `"`,
		`  end if`,
		`  try`,
		`    select targetWindow`,
		`    select targetTab`,
		`    select targetSession`,
		`  on error`,
		`  end try`,
		`  tell targetSession`,
		`    ` + writeLine,
		`  end tell`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// SendText writes text into the iTerm2 session identified by sessionID.
// Search order matches Contents: ITERM2_APP_PATH (if usable), ~/Applications,
// /Applications — skipping installs that are not running.
func SendText(sessionID, text string, opts SendTextOptions, cfg *SendTextConfig) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return fmt.Errorf("session id is required")
	}
	uuid := SessionUUID(sessionID)
	if strings.TrimSpace(uuid) == "" {
		return fmt.Errorf("session id is required")
	}

	execFn := sendTextOsascript
	if cfg != nil && cfg.Exec != nil {
		execFn = cfg.Exec
	}

	apps := sendTextSearchApps(cfg)
	var lastNotFound error
	for _, app := range apps {
		script := BuildSendTextScript(uuid, text, opts, app.Path)
		err := execFn(script)
		if err != nil {
			if isSessionNotFound(err) {
				lastNotFound = err
				continue
			}
			return err
		}
		return nil
	}
	if lastNotFound != nil {
		return fmt.Errorf("%w: %s", ErrSessionNotFound, uuid)
	}
	return fmt.Errorf("%w: %s", ErrSessionNotFound, uuid)
}

func sendTextSearchApps(cfg *SendTextConfig) []ContentsApp {
	if cfg != nil && cfg.Apps != nil {
		return cfg.Apps
	}
	cc := &ContentsConfig{}
	if cfg != nil {
		cc.Getenv = cfg.Getenv
		cc.Home = cfg.Home
		cc.IsApp = cfg.IsApp
		cc.Running = cfg.Running
	}
	return contentsSearchApps(cc)
}

func sendTextOsascript(script string) error {
	_, err := contentsOsascript(script)
	return err
}
