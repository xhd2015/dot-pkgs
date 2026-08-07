//go:build darwin

package chrome

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// nilCtx is used for timeout-bounded helpers without a parent cancel signal.
func nilCtx() context.Context { return context.Background() }

func loadUnpacked(ctx context.Context, opts LoadUnpackedOpts) (LoadUnpackedResult, error) {
	opts, err := normalizeOpts(opts)
	if err != nil {
		return LoadUnpackedResult{}, err
	}
	if err := ctx.Err(); err != nil {
		return LoadUnpackedResult{}, err
	}

	var res LoadUnpackedResult
	shotN := 0

	if hint := MultiInstanceHint(); hint != "" {
		res.MultiInstanceWarned = true
		stepf(opts.Stderr, "warning: %s", hint)
	}
	if opts.ScreenshotDir != "" {
		stepf(opts.Stdout, "  shots     %s", opts.ScreenshotDir)
	}

	stepf(opts.Stdout, "  app       %s", opts.AppName)
	stepf(opts.Stdout, "  extension %s", opts.ExtensionDir)
	stepf(opts.Stdout, "  mode      ui (no Chrome flags, default profile)")

	if err := launchChromeBare(ctx, opts.AppName); err != nil {
		return res, err
	}
	stepf(opts.Stdout, "  step      launch_chrome        ok  (bare, no flags)")
	snapStep(opts, &res, "01-after-launch", &shotN)
	// Wait for session restore so later navigation is not overwritten.
	sleepCtx(ctx, 2000*time.Millisecond)
	snapStep(opts, &res, "02-after-restore-wait", &shotN)

	// Single window: navigate front tab to chrome://extensions and verify URL.
	// Never click Developer mode until we confirm we are on the extensions page
	// (otherwise coordinate fallback can "misclick" restored tabs e.g. Docs).
	how, err := ensureOnExtensionsPage(ctx, opts.AppName)
	if err != nil {
		snapStep(opts, &res, "03-open-extensions-FAIL", &shotN)
		return res, err
	}
	stepf(opts.Stdout, "  step      open_extensions      ok  (%s, %s)", ExtensionsURL, how)
	sleepCtx(ctx, 1500*time.Millisecond)
	snapStep(opts, &res, "03-after-open-extensions", &shotN)

	// Probe UI state for dry-run / dump / load path (light query — avoid entire contents).
	dev, loadBtn, listed, probeErr := probeExtensionsUI(ctx, opts.AppName, opts.VerifyName)
	if probeErr != nil {
		stepf(opts.Stderr, "warning: could not probe extensions UI: %v", probeErr)
	}
	res.DeveloperModeVisible = dev
	res.LoadUnpackedVisible = loadBtn
	res.ExtensionListed = listed
	stepf(opts.Stdout, "  step      probe                dev=%s load_unpacked=%s listed=%s", yn(dev), yn(loadBtn), yn(listed))
	snapStep(opts, &res, "04-after-probe", &shotN)

	if opts.DumpTree {
		tree, err := dumpUITree(ctx, opts.AppName)
		if err != nil {
			return res, fmt.Errorf("chrome: dump-tree: %w", err)
		}
		if _, err := opts.Stderr.Write([]byte(tree)); err != nil {
			return res, err
		}
		if !strings.HasSuffix(tree, "\n") {
			_, _ = opts.Stderr.Write([]byte("\n"))
		}
		snapStep(opts, &res, "05-after-dump-tree", &shotN)
		return res, nil
	}

	if opts.DryRun {
		stepf(opts.Stdout, "  step      dry_run             ok  developer_mode_ctrl=%s", yn(dev))
		stepf(opts.Stdout, "  step      dry_run             ok  load_unpacked=%s", yn(loadBtn))
		stepf(opts.Stdout, "  step      dry_run             ok  browser_agent_listed=%s", yn(listed))
		stepf(opts.Stdout, "dry-run complete (no clicks)")
		snapStep(opts, &res, "05-dry-run-done", &shotN)
		return res, nil
	}

	// Full load path.
	if err := ensureDeveloperMode(ctx, opts); err != nil {
		snapStep(opts, &res, "05-developer-mode-FAIL", &shotN)
		return res, err
	}
	snapStep(opts, &res, "05-after-developer-mode", &shotN)
	// Re-probe after toggle.
	_, loadBtn, _, _ = probeExtensionsUI(ctx, opts.AppName, opts.VerifyName)
	res.LoadUnpackedVisible = loadBtn

	if err := clickLoadUnpacked(ctx, opts.AppName); err != nil {
		snapStep(opts, &res, "06-load-unpacked-FAIL", &shotN)
		return res, err
	}
	stepf(opts.Stdout, "  step      load_unpacked        click")
	snapStep(opts, &res, "06-after-load-unpacked-click", &shotN)

	if err := waitOpenSheet(ctx, opts.AppName, opts.DialogTimeout); err != nil {
		snapStep(opts, &res, "07-open-dialog-FAIL", &shotN)
		return res, err
	}
	stepf(opts.Stdout, "  step      open_dialog          ok")
	snapStep(opts, &res, "07-after-open-dialog", &shotN)

	if err := pickFolderViaKeystrokes(ctx, opts.AppName, opts.ExtensionDir); err != nil {
		snapStep(opts, &res, "08-pick-folder-FAIL", &shotN)
		return res, err
	}
	stepf(opts.Stdout, "  step      pick_folder          ok  (%s)", opts.ExtensionDir)
	snapStep(opts, &res, "08-after-pick-folder", &shotN)

	sleepCtx(ctx, time.Second)
	deadline := time.Now().Add(opts.VerifyTimeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		// AX first: System Events often misses "Browser Agent" cards on the web UI.
		if axExistsNamed(opts.AppName, opts.VerifyName) {
			res.ExtensionListed = true
			res.Loaded = true
			stepf(opts.Stdout, "  status    loaded  (%s visible via AX)", opts.VerifyName)
			snapStep(opts, &res, "09-loaded", &shotN)
			maybeRemoveOlder(ctx, opts, &res)
			snapStep(opts, &res, "10-after-remove-older", &shotN)
			return res, nil
		}
		_, _, listed, _ = probeExtensionsUI(ctx, opts.AppName, opts.VerifyName)
		if listed {
			res.ExtensionListed = true
			res.Loaded = true
			stepf(opts.Stdout, "  status    loaded  (%s visible on extensions page)", opts.VerifyName)
			snapStep(opts, &res, "09-loaded", &shotN)
			maybeRemoveOlder(ctx, opts, &res)
			snapStep(opts, &res, "10-after-remove-older", &shotN)
			return res, nil
		}
		sleepCtx(ctx, 350*time.Millisecond)
	}
	res.SubmittedUnknown = true
	stepf(opts.Stderr, "warning: folder selected; could not confirm extension card on page")
	stepf(opts.Stdout, "  status    submitted_unknown")
	snapStep(opts, &res, "09-submitted-unknown", &shotN)
	// Still attempt cleanup after a submitted load (best-effort).
	maybeRemoveOlder(ctx, opts, &res)
	snapStep(opts, &res, "10-after-remove-older", &shotN)
	return res, nil
}

// maybeRemoveOlder runs post-load cleanup when enabled and load succeeded.
func maybeRemoveOlder(ctx context.Context, opts LoadUnpackedOpts, res *LoadUnpackedResult) {
	if res == nil || !removeOlderEnabled(opts) {
		return
	}
	if !res.Loaded && !res.SubmittedUnknown {
		return
	}
	res.RemoveOlderAttempted = true
	// Print before work so a multi-second scan does not look like a hang.
	stepf(opts.Stdout, "  step      remove_older         scanning...")
	n, err := removeOlderExtensions(ctx, opts)
	res.RemovedOlder = n
	if err != nil {
		stepf(opts.Stderr, "warning: could not remove older extensions: %v", err)
		return
	}
	if n > 0 {
		stepf(opts.Stdout, "  step      remove_older         removed %d same-name card(s)", n)
	} else {
		stepf(opts.Stdout, "  step      remove_older         none")
	}
}

func yn(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func sleepCtx(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
	case <-t.C:
	}
}

func runOSAscript(ctx context.Context, source string, timeout time.Duration) (stdout, stderr string, err error) {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, "osascript", "-e", source)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return strings.TrimSpace(outBuf.String()), strings.TrimSpace(errBuf.String()), err
}

func launchChromeBare(ctx context.Context, appName string) error {
	cmd := exec.CommandContext(ctx, "open", "-a", appName)
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(errBuf.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("chrome: could not open %q: %s", appName, msg)
	}
	return nil
}

// ensureOnExtensionsPage puts chrome://extensions in the front window without
// always creating a second window. Always verifies the front tab URL.
// how: "reused"|"created"|"new-tab".
func ensureOnExtensionsPage(ctx context.Context, appName string) (how string, err error) {
	app := escapeAS(appName)
	// Navigate via AppleScript; prefer reusing one window. Open a *new tab* for
	// extensions when the active tab is some restored page (Docs, etc.) so we do
	// not fight session restore on that tab.
	script := fmt.Sprintf(`
tell application "%s"
  activate
  set didCreate to false
  if (count of windows) = 0 then
    make new window
    set didCreate to true
  end if
  set u to ""
  try
    set u to URL of active tab of front window as text
  end try
  if u starts with "chrome://extensions" then
    return "reused"
  end if
  -- Active tab is something else (often session-restored Docs). Open a new tab
  -- for extensions instead of set-URL races / later tab-strip misclicks.
  try
    tell front window
      set t to make new tab with properties {URL:"%s"}
      set active tab index to (index of t)
    end tell
    delay 0.7
  on error
    open location "%s"
    delay 0.7
  end try
  set u to ""
  try
    set u to URL of active tab of front window as text
  end try
  if u does not start with "chrome://extensions" then
    -- Force set URL on active tab as last resort
    try
      set URL of active tab of front window to "%s"
      delay 0.6
      set u to URL of active tab of front window as text
    end try
  end if
  if u does not start with "chrome://extensions" then
    return "fail:" & u
  end if
  if didCreate then
    return "created"
  end if
  return "new-tab"
end tell
`, app, ExtensionsURL, ExtensionsURL, ExtensionsURL)
	out, errOut, runErr := runOSAscript(ctx, script, 45*time.Second)
	if runErr != nil {
		if errOut != "" {
			return "", fmt.Errorf("chrome: open %s: %s", ExtensionsURL, errOut)
		}
		return "", fmt.Errorf("chrome: open %s: %w", ExtensionsURL, runErr)
	}
	out = strings.TrimSpace(out)
	if strings.HasPrefix(out, "fail:") {
		return "", fmt.Errorf("chrome: front tab is not %s (got %s) — refuse to click (avoids misclick on restored pages)", ExtensionsURL, strings.TrimPrefix(out, "fail:"))
	}
	if out == "" {
		return "reused", nil
	}
	return out, nil
}

// frontTabURL returns the active tab URL of the front Chrome window (best-effort).
func frontTabURL(ctx context.Context, appName string) string {
	app := escapeAS(appName)
	out, _, err := runOSAscript(ctx, fmt.Sprintf(`
tell application "%s"
  try
    return URL of active tab of front window as text
  on error
    return ""
  end try
end tell
`, app), 10*time.Second)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out)
}

// requireExtensionsURL errors unless front tab is chrome://extensions.
func requireExtensionsURL(ctx context.Context, appName string) error {
	u := frontTabURL(ctx, appName)
	if strings.HasPrefix(u, "chrome://extensions") {
		return nil
	}
	// One more navigation attempt
	if _, err := ensureOnExtensionsPage(ctx, appName); err != nil {
		return err
	}
	u = frontTabURL(ctx, appName)
	if strings.HasPrefix(u, "chrome://extensions") {
		return nil
	}
	return fmt.Errorf("chrome: expected front tab %s, got %q (refusing clicks)", ExtensionsURL, u)
}

// probeExtensionsUI uses light System Events queries (no entire contents — that
// often gets killed on large extensions pages).
func probeExtensionsUI(ctx context.Context, appName, verifyName string) (dev, loadBtn, listed bool, err error) {
	app := escapeAS(appName)
	name := escapeAS(verifyName)
	script := fmt.Sprintf(`
tell application "System Events"
  tell process "%s"
    set frontmost to true
    set devOn to false
    set loadOn to false
    set listedOn to false
    try
      if exists (first checkbox whose name is "Developer mode") then set devOn to true
    end try
    try
      if exists (first UI element whose name is "Developer mode") then set devOn to true
    end try
    try
      if exists (first button whose name is "Load unpacked") then set loadOn to true
    end try
    try
      if exists (first UI element whose name is "Load unpacked") then set loadOn to true
    end try
    try
      if exists (first UI element whose name is "%s") then set listedOn to true
    end try
    try
      if exists (first static text whose value contains "%s") then set listedOn to true
    end try
    return (devOn as text) & "," & (loadOn as text) & "," & (listedOn as text)
  end tell
end tell
`, app, name, name)
	out, errOut, runErr := runOSAscript(ctx, script, 20*time.Second)
	if runErr != nil {
		if errOut != "" {
			return false, false, false, fmt.Errorf("%s", errOut)
		}
		return false, false, false, runErr
	}
	parts := strings.Split(out, ",")
	for len(parts) < 3 {
		parts = append(parts, "false")
	}
	return parts[0] == "true", parts[1] == "true", parts[2] == "true", nil
}

func ensureDeveloperMode(ctx context.Context, opts LoadUnpackedOpts) error {
	if err := requireExtensionsURL(ctx, opts.AppName); err != nil {
		return err
	}
	// Prefer AX existence: System Events often false-negatives on Chrome web UI
	// even when Load unpacked is clearly visible (confirmed via screenshots).
	if axExistsNamed(opts.AppName, "Load unpacked") {
		stepf(opts.Stdout, "  step      developer_mode       already_on  (AX Load unpacked)")
		return nil
	}
	_, loadBtn, _, _ := probeExtensionsUI(ctx, opts.AppName, opts.VerifyName)
	if loadBtn {
		stepf(opts.Stdout, "  step      developer_mode       already_on")
		return nil
	}
	// AX + Quartz click at real control frame (same as Python loader).
	// Never blind window-relative coordinates (those hit the tab strip → Docs).
	if err := axClickNamed(opts.AppName, "Developer mode"); err != nil {
		return fmt.Errorf("chrome: enable Developer mode: %w", err)
	}
	stepf(opts.Stdout, "  step      developer_mode       toggled-ax")
	// Wait for Load unpacked (AX first — SE probe is flaky here).
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		if u := frontTabURL(ctx, opts.AppName); u != "" && !strings.HasPrefix(u, "chrome://extensions") {
			return fmt.Errorf("chrome: left %s during developer mode (now %q)", ExtensionsURL, u)
		}
		if axExistsNamed(opts.AppName, "Load unpacked") {
			stepf(opts.Stdout, "  step      developer_mode       on  (AX Load unpacked visible)")
			return nil
		}
		_, loadBtn, _, _ = probeExtensionsUI(ctx, opts.AppName, opts.VerifyName)
		if loadBtn {
			stepf(opts.Stdout, "  step      developer_mode       on  (Load unpacked visible)")
			return nil
		}
		sleepCtx(ctx, 350*time.Millisecond)
	}
	return fmt.Errorf("chrome: Developer mode toggle did not reveal Load unpacked")
}

func clickLoadUnpacked(ctx context.Context, appName string) error {
	if err := requireExtensionsURL(ctx, appName); err != nil {
		return err
	}
	// Prefer AX + Quartz (reliable on Chrome web UI); System Events as backup.
	if err := axClickNamed(appName, "Load unpacked"); err == nil {
		return nil
	}
	app := escapeAS(appName)
	script := fmt.Sprintf(`
tell application "System Events"
  tell process "%s"
    set frontmost to true
    delay 0.25
    try
      click (first button whose name is "Load unpacked")
      return "ok"
    end try
    try
      click (first UI element whose name is "Load unpacked")
      return "ok"
    end try
    return "err:Load unpacked button not found"
  end tell
end tell
`, app)
	out, errOut, err := runOSAscript(ctx, script, 20*time.Second)
	if err != nil || strings.HasPrefix(out, "err:") {
		msg := out
		if errOut != "" {
			msg = errOut
		}
		if msg == "" && err != nil {
			msg = err.Error()
		}
		return fmt.Errorf("chrome: click Load unpacked: %s", msg)
	}
	return nil
}

func waitOpenSheet(ctx context.Context, appName string, timeout time.Duration) error {
	app := escapeAS(appName)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := ctx.Err(); err != nil {
			return err
		}
		script := fmt.Sprintf(`
tell application "System Events"
  tell process "%s"
    try
      return (count of sheets of front window) as text
    on error
      return "0"
    end try
  end tell
end tell
`, app)
		out, _, err := runOSAscript(ctx, script, 10*time.Second)
		if err == nil && out != "" && out != "0" {
			return nil
		}
		sleepCtx(ctx, 300*time.Millisecond)
	}
	return fmt.Errorf("chrome: open folder dialog did not appear after Load unpacked")
}

func pickFolderViaKeystrokes(ctx context.Context, appName, extensionDir string) error {
	app := escapeAS(appName)
	path := escapeAS(extensionDir)
	script := fmt.Sprintf(`
tell application "%s" to activate
delay 0.3
tell application "System Events"
  tell process "%s"
    set frontmost to true
    delay 0.2
    keystroke "g" using {command down, shift down}
    delay 0.7
    keystroke "%s"
    delay 0.25
    keystroke return
    delay 0.9
    try
      click button "Open" of sheet 1 of front window
    on error
      try
        click button "Open" of front window
      on error
        keystroke return
      end try
    end try
  end tell
end tell
return "ok"
`, app, app, path)
	_, errOut, err := runOSAscript(ctx, script, 45*time.Second)
	if err != nil {
		if errOut != "" {
			return fmt.Errorf("chrome: folder picker: %s", errOut)
		}
		return fmt.Errorf("chrome: folder picker: %w", err)
	}
	return nil
}

func dumpUITree(ctx context.Context, appName string) (string, error) {
	app := escapeAS(appName)
	// Limited dump: name + role of interesting UI elements (stderr destination chosen by caller).
	script := fmt.Sprintf(`
tell application "System Events"
  tell process "%s"
    set frontmost to true
    set out to ""
    try
      set uiElems to entire contents of front window
      set n to 0
      repeat with e in uiElems
        set n to n + 1
        if n > 400 then exit repeat
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
        set keep to false
        if r is in {"button", "checkbox", "static text", "window", "sheet", "text field", "group", "web area"} then set keep to true
        set low to nm
        considering case
          -- AppleScript has no lower; use contains on known tokens
        end considering
        if nm contains "Develop" or nm contains "Load" or nm contains "unpack" or nm contains "Browser" or nm contains "xtension" then set keep to true
        if keep and (nm is not "") then
          set out to out & r & " name=" & nm & linefeed
        end if
      end repeat
    on error errMsg
      set out to "error: " & errMsg & linefeed
    end try
    return out
  end tell
end tell
`, app)
	out, errOut, err := runOSAscript(ctx, script, 60*time.Second)
	if err != nil {
		if errOut != "" {
			return "", fmt.Errorf("%s", errOut)
		}
		return "", err
	}
	if out == "" {
		return "(empty UI tree)\n", nil
	}
	return out, nil
}
