//go:build darwin

package chrome

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// removeOlder timeouts: full "entire contents" walks on chrome://extensions are
// slow; prefer AX/light early-out, and soft-skip rather than hanging near 40s.
const (
	removeOlderScanTimeout  = 8 * time.Second
	removeOlderClickTimeout = 10 * time.Second
)

// removeOlderExtensions best-effort removes extension cards named VerifyName
// that do not match the just-loaded VerifyVersion (when known).
//
// Strategy:
//  1. Fast path: AX name-count (or light SE) — if ≤1 same-name hit, done (no full tree).
//  2. Else System Events "entire contents" scan/click with short timeouts.
//  3. Soft-fail on timeout: return error for a short warning; do not block long.
//
// Soft failures return (partial count, err) or (0, err); callers warn only.
func removeOlderExtensions(ctx context.Context, opts LoadUnpackedOpts) (int, error) {
	app := escapeAS(opts.AppName)
	name := escapeAS(opts.VerifyName)
	keepVer := strings.TrimSpace(opts.VerifyVersion)
	keepVerAS := escapeAS(keepVer)

	removed := 0
	const maxRounds = 8
	for round := 0; round < maxRounds; round++ {
		if err := ctx.Err(); err != nil {
			return removed, err
		}

		// Fast path: AX count avoids System Events entire-contents on the common
		// single-card case (often many seconds otherwise).
		if hits, ok := countSameNameCardsFast(opts); ok && hits <= 1 {
			return removed, nil
		}

		// Snapshot: count same-name hits (heavy; short timeout + soft skip).
		script := fmt.Sprintf(`
tell application "System Events"
  tell process "%s"
    set frontmost to true
    set nameHits to 0
    set extName to "%s"
    try
      set uiElems to entire contents of front window
      repeat with e in uiElems
        try
          set nm to name of e as text
        on error
          set nm to ""
        end try
        if nm is extName then set nameHits to nameHits + 1
      end repeat
    end try
    return nameHits as text
  end tell
end tell
`, app, name)

		out, errOut, err := runOSAscript(ctx, script, removeOlderScanTimeout)
		if err != nil {
			if removed > 0 {
				return removed, fmt.Errorf("scan after %d remove(s): %v %s", removed, err, errOut)
			}
			if isTimeoutErr(err) {
				return 0, fmt.Errorf("scan timed out after %s (skipped)", removeOlderScanTimeout)
			}
			return 0, fmt.Errorf("scan extensions UI: %v %s", err, errOut)
		}
		nameHits := 0
		fmt.Sscanf(strings.TrimSpace(out), "%d", &nameHits)
		// Stop if at most one card with our name.
		if nameHits <= 1 {
			return removed, nil
		}

		// Click one Remove associated with a non-keep card.
		clickScript := fmt.Sprintf(`
tell application "System Events"
  tell process "%s"
    set frontmost to true
    delay 0.15
    set extName to "%s"
    set keepVer to "%s"
    set clicked to false
    try
      set uiElems to entire contents of front window
      set lastWasName to false
      set lastWasOtherVer to false
      repeat with e in uiElems
        try
          set nm to name of e as text
        on error
          set nm to ""
        end try
        try
          set r to role of e as text
        on error
          set r to ""
        end try
        if nm is extName then
          set lastWasName to true
          set lastWasOtherVer to false
        else if lastWasName and nm is not "" then
          -- next labels after name: often version
          if keepVer is not "" and nm is keepVer then
            set lastWasName to false
            set lastWasOtherVer to false
          else if keepVer is not "" and nm is not keepVer then
            try
              set c1 to character 1 of nm
              if c1 is in "0123456789" and nm contains "." then
                set lastWasOtherVer to true
              end if
            end try
          else if keepVer is "" then
            -- no keep version: treat second+ name cards' Remove as older
            set lastWasOtherVer to true
          end if
        end if
        if (lastWasOtherVer or (lastWasName and keepVer is "")) and (nm is "Remove" or nm is "Remove from Chrome") then
          try
            click e
            set clicked to true
            exit repeat
          end try
        end if
        if nm is "Remove" or nm is "Remove from Chrome" then
          -- fallback: if we already saw ext name and another version this pass
          if lastWasOtherVer then
            try
              click e
              set clicked to true
              exit repeat
            end try
          end if
        end if
      end repeat
    end try
    if not clicked then
      -- last resort: click first Remove button after any second name hit (fragile)
      try
        set nameCount to 0
        set uiElems to entire contents of front window
        repeat with e in uiElems
          try
            set nm to name of e as text
          on error
            set nm to ""
          end try
          if nm is extName then set nameCount to nameCount + 1
          if nameCount is greater than or equal to 2 and (nm is "Remove" or nm is "Remove from Chrome") then
            click e
            set clicked to true
            exit repeat
          end if
        end repeat
      end try
    end if
    if clicked then
      delay 0.4
      -- Confirm dialog if present
      try
        click button "Remove" of sheet 1 of front window
      on error
        try
          click button "Remove" of front window
        on error
          try
            keystroke return
          end try
        end try
      end try
      return "clicked"
    end if
    return "none"
  end tell
end tell
`, app, name, keepVerAS)

		cout, cerr, cerrr := runOSAscript(ctx, clickScript, removeOlderClickTimeout)
		if cerrr != nil {
			if removed > 0 {
				return removed, fmt.Errorf("click Remove after %d: %v %s", removed, cerrr, cerr)
			}
			if isTimeoutErr(cerrr) {
				return 0, fmt.Errorf("click timed out after %s (skipped)", removeOlderClickTimeout)
			}
			return 0, fmt.Errorf("click Remove: %v %s", cerrr, cerr)
		}
		if strings.TrimSpace(cout) != "clicked" {
			// Nothing to click — stop.
			return removed, nil
		}
		removed++
		sleepCtx(ctx, 800*time.Millisecond)
	}
	return removed, nil
}

// countSameNameCardsFast returns (hits, true) when a cheap signal is available.
// hits ≤ 1 means no multi-card cleanup is needed.
// ok=false means fall through to a bounded heavy SE scan.
func countSameNameCardsFast(opts LoadUnpackedOpts) (hits int, ok bool) {
	name := strings.TrimSpace(opts.VerifyName)
	if name == "" {
		return 0, true
	}
	if n := axCountNamed(opts.AppName, name); n >= 0 {
		return n, true
	}
	// AX unavailable: light SE existence only distinguishes 0 vs ≥1, not multi.
	// If nothing listed under the name, treat as none; otherwise escalate.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_, _, listed, err := probeExtensionsUI(ctx, opts.AppName, name)
	if err != nil {
		return 0, false
	}
	if !listed {
		return 0, true
	}
	// Present but multi-count unknown without AX — escalate to bounded SE scan.
	return 0, false
}
