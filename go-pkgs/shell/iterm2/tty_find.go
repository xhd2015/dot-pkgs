package iterm2

import (
	"fmt"
	"strings"
)

// BuildFindByTTYScript returns AppleScript that scans windows/tabs/sessions and
// emits only rows whose tty matches one of ttys (after /dev/ normalization).
// Output shape matches BuildSessionListScript so ParseSessionListOutput works:
//
//	WindowID\tWindowName\tTabIndex\tSessionIndex\tSessionID\tTTY\tName
//
// Uses `repeat with … in` enumeration (faster than index snapshots for this
// targeted scan). Soft-skips per-item errors so mid-scan churn does not abort.
func BuildFindByTTYScript(ttys []string) string {
	want := uniqueNormalizedTTYs(ttys)
	sep := fieldSepAS
	lines := []string{
		tellHeaderResolved(),
		`  set fieldSep to ` + sep,
		`  set outLines to {}`,
		`  set wantTTYs to {` + appleScriptStringList(want) + `}`,
		`  repeat with aWindow in windows`,
		`    try`,
		`      set windowID to ""`,
		`      set windowName to ""`,
		`      try`,
		`        set windowID to id of aWindow as string`,
		`      end try`,
		`      try`,
		`        set windowName to name of aWindow as string`,
		`      end try`,
		`      set ti to 0`,
		`      repeat with aTab in tabs of aWindow`,
		`        set ti to ti + 1`,
		`        try`,
		`          set si to 0`,
		`          repeat with aSession in sessions of aTab`,
		`            set si to si + 1`,
		`            try`,
		`              set sessionTTY to ""`,
		`              try`,
		`                set sessionTTY to tty of aSession`,
		`              end try`,
		`              set normTTY to sessionTTY`,
		`              if normTTY does not start with "/dev/" and normTTY is not "" then`,
		`                set normTTY to "/dev/" & normTTY`,
		`              end if`,
		`              if wantTTYs contains normTTY or wantTTYs contains sessionTTY then`,
		`                set sessionID to ""`,
		`                set sessionName to ""`,
		`                try`,
		`                  set sessionID to id of aSession as string`,
		`                end try`,
		`                try`,
		`                  set sessionName to name of aSession as string`,
		`                end try`,
		`                set end of outLines to windowID & fieldSep & windowName & fieldSep & (ti as string) & fieldSep & (si as string) & fieldSep & sessionID & fieldSep & normTTY & fieldSep & sessionName`,
		`              end if`,
		`            end try`,
		`          end repeat`,
		`        end try`,
		`      end repeat`,
		`    end try`,
		`  end repeat`,
		`  set AppleScript's text item delimiters to linefeed`,
		`  set joined to outLines as text`,
		`  set AppleScript's text item delimiters to ""`,
		`  return joined`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// FindSessionsByTTY lists iTerm sessions whose TTY matches any of ttys.
// Empty ttys yields an empty slice without calling osascript.
func FindSessionsByTTY(ttys []string) ([]SessionRef, error) {
	want := uniqueNormalizedTTYs(ttys)
	if len(want) == 0 {
		return []SessionRef{}, nil
	}
	return runSessionListScript(BuildFindByTTYScript(want))
}

// BuildContentsByTTYScript returns AppleScript that returns contents of the
// first session whose tty matches tty (normalized). Early-exits on hit.
func BuildContentsByTTYScript(tty, appPath string) string {
	norm := NormalizeTTY(tty)
	escaped := EscapePathForAppleScript(norm)
	bare := strings.TrimPrefix(norm, "/dev/")
	bareEsc := EscapePathForAppleScript(bare)
	header := TellApplicationHeader(appPath)
	lines := []string{
		header,
		`  set wantTTY to "` + escaped + `"`,
		`  set wantBare to "` + bareEsc + `"`,
		`  repeat with aWindow in windows`,
		`    try`,
		`      repeat with aTab in tabs of aWindow`,
		`        try`,
		`          repeat with aSession in sessions of aTab`,
		`            try`,
		`              set sessionTTY to tty of aSession`,
		`              set normTTY to sessionTTY`,
		`              if normTTY does not start with "/dev/" and normTTY is not "" then`,
		`                set normTTY to "/dev/" & normTTY`,
		`              end if`,
		`              if normTTY is wantTTY or sessionTTY is wantBare or sessionTTY is wantTTY then`,
		`                return contents of aSession`,
		`              end if`,
		`            end try`,
		`          end repeat`,
		`        end try`,
		`      end repeat`,
		`    end try`,
		`  end repeat`,
		`  error "tty not found: ` + escaped + `"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// ContentsByTTY dumps visible text for the first iTerm session on tty.
// Same app search order as Contents. Early-exits inside AppleScript on hit.
func ContentsByTTY(tty string, cfg *ContentsConfig) (ContentsResult, error) {
	tty = NormalizeTTY(tty)
	if tty == "" {
		return ContentsResult{}, fmt.Errorf("tty is required")
	}

	execFn := contentsOsascript
	if cfg != nil && cfg.Exec != nil {
		execFn = cfg.Exec
	}

	apps := contentsSearchApps(cfg)
	var lastNotFound error
	for _, app := range apps {
		script := BuildContentsByTTYScript(tty, app.Path)
		out, err := execFn(script)
		if err != nil {
			if isTTYNotFound(err) || isSessionNotFound(err) {
				lastNotFound = err
				continue
			}
			return ContentsResult{}, err
		}
		return ContentsResult{
			App:      app.Canonical,
			Contents: strings.TrimRight(out, "\r\n"),
		}, nil
	}
	if lastNotFound != nil {
		return ContentsResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, tty)
	}
	return ContentsResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, tty)
}

// CaptureByTTYResult is one early-exit pane dump keyed by TTY.
type CaptureByTTYResult struct {
	Ref      SessionRef
	Contents string
	App      string
}

// metaPrefix marks the first line of CaptureByTTY osascript stdout.
const captureByTTYMetaPrefix = "#meta\t"

// BuildCaptureByTTYScript early-exits on the first matching tty and returns:
//
//	#meta\tWindowID\tWindowName\tTabIndex\tSessionIndex\tSessionID\tTTY\tName
//	<contents…>
func BuildCaptureByTTYScript(tty, appPath string) string {
	norm := NormalizeTTY(tty)
	escaped := EscapePathForAppleScript(norm)
	bare := strings.TrimPrefix(norm, "/dev/")
	bareEsc := EscapePathForAppleScript(bare)
	header := TellApplicationHeader(appPath)
	sep := fieldSepAS
	lines := []string{
		header,
		`  set fieldSep to ` + sep,
		`  set wantTTY to "` + escaped + `"`,
		`  set wantBare to "` + bareEsc + `"`,
		`  repeat with aWindow in windows`,
		`    try`,
		`      set windowID to ""`,
		`      set windowName to ""`,
		`      try`,
		`        set windowID to id of aWindow as string`,
		`      end try`,
		`      try`,
		`        set windowName to name of aWindow as string`,
		`      end try`,
		`      set ti to 0`,
		`      repeat with aTab in tabs of aWindow`,
		`        set ti to ti + 1`,
		`        try`,
		`          set si to 0`,
		`          repeat with aSession in sessions of aTab`,
		`            set si to si + 1`,
		`            try`,
		`              set sessionTTY to tty of aSession`,
		`              set normTTY to sessionTTY`,
		`              if normTTY does not start with "/dev/" and normTTY is not "" then`,
		`                set normTTY to "/dev/" & normTTY`,
		`              end if`,
		`              if normTTY is wantTTY or sessionTTY is wantBare or sessionTTY is wantTTY then`,
		`                set sessionID to ""`,
		`                set sessionName to ""`,
		`                try`,
		`                  set sessionID to id of aSession as string`,
		`                end try`,
		`                try`,
		`                  set sessionName to name of aSession as string`,
		`                end try`,
		`                set meta to "#meta" & fieldSep & windowID & fieldSep & windowName & fieldSep & (ti as string) & fieldSep & (si as string) & fieldSep & sessionID & fieldSep & normTTY & fieldSep & sessionName`,
		`                set body to contents of aSession`,
		`                return meta & linefeed & body`,
		`              end if`,
		`            end try`,
		`          end repeat`,
		`        end try`,
		`      end repeat`,
		`    end try`,
		`  end repeat`,
		`  error "tty not found: ` + escaped + `"`,
		`end tell`,
	}
	return strings.Join(lines, "\n")
}

// ParseCaptureByTTYOutput splits CaptureByTTY osascript stdout into ref + body.
func ParseCaptureByTTYOutput(stdout string) (CaptureByTTYResult, error) {
	stdout = strings.TrimRight(stdout, "\r")
	if stdout == "" {
		return CaptureByTTYResult{}, fmt.Errorf("empty capture output")
	}
	line, rest, ok := strings.Cut(stdout, "\n")
	line = strings.TrimRight(line, "\r")
	if !ok || !strings.HasPrefix(line, captureByTTYMetaPrefix) {
		return CaptureByTTYResult{}, fmt.Errorf("missing capture meta line")
	}
	refs, err := ParseSessionListOutput(strings.TrimPrefix(line, "#meta\t"))
	if err != nil {
		return CaptureByTTYResult{}, err
	}
	if len(refs) != 1 {
		return CaptureByTTYResult{}, fmt.Errorf("capture meta: want 1 ref, got %d", len(refs))
	}
	return CaptureByTTYResult{
		Ref:      refs[0],
		Contents: strings.TrimRight(rest, "\r\n"),
	}, nil
}

// CaptureByTTY early-exits on tty and returns session meta + contents in one
// AppleScript. Same app search order as Contents.
func CaptureByTTY(tty string, cfg *ContentsConfig) (CaptureByTTYResult, error) {
	tty = NormalizeTTY(tty)
	if tty == "" {
		return CaptureByTTYResult{}, fmt.Errorf("tty is required")
	}

	execFn := contentsOsascript
	if cfg != nil && cfg.Exec != nil {
		execFn = cfg.Exec
	}

	apps := contentsSearchApps(cfg)
	var lastNotFound error
	for _, app := range apps {
		script := BuildCaptureByTTYScript(tty, app.Path)
		out, err := execFn(script)
		if err != nil {
			if isTTYNotFound(err) || isSessionNotFound(err) {
				lastNotFound = err
				continue
			}
			return CaptureByTTYResult{}, err
		}
		res, parseErr := ParseCaptureByTTYOutput(out)
		if parseErr != nil {
			return CaptureByTTYResult{}, parseErr
		}
		res.App = app.Canonical
		return res, nil
	}
	if lastNotFound != nil {
		return CaptureByTTYResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, tty)
	}
	return CaptureByTTYResult{}, fmt.Errorf("%w: %s", ErrSessionNotFound, tty)
}

func isTTYNotFound(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "tty not found")
}

func uniqueNormalizedTTYs(ttys []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, t := range ttys {
		n := NormalizeTTY(t)
		if n == "" {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func appleScriptStringList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, `"`+EscapePathForAppleScript(v)+`"`)
	}
	return strings.Join(parts, ", ")
}
