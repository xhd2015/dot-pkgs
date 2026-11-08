package reclaim

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// Options configures the clear-tmp pipeline (injectable for L2 tests).
type Options struct {
	Home                    string
	DryRun                  bool
	IncludeIOSSim           bool // true when --ios-sim: opt-in iOS Simulator runtime delete
	IncludeIOSSimAssets     bool // true when --ios-sim-assets: wipe MobileAsset cache (may sudo)
	IncludeMSPlaywright     bool // true when --ms-playwright: wipe Library/Caches/ms-playwright
	IncludeOpenCodeSnapshot bool // true when --opencode-snapshot: wipe .local/share/opencode/snapshot
	Aggressive              bool // true when --aggressive-all: also run TierAggressive steps
	UseColor                bool
	Stdout                  io.Writer
	Stderr                  io.Writer
	PathEnv                 string // if non-empty: exclusive PATH for brew/npm/xcrun children
	SkipVolume              bool

	// IOSSimAssetRoot overrides the MobileAsset path (tests only). Empty → system default.
	IOSSimAssetRoot string
	// Geteuid injects euid (tests). Nil → os.Geteuid.
	Geteuid func() int
	// IsTerminal injects TTY detection for sudo prompt (tests). Nil → /dev/tty or stdin.
	IsTerminal func() bool
	// SudoRun injects elevation (tests). Nil → interactive sudo rm of allowlisted path.
	// When set, it should perform the privileged wipe of Options.IOSSimAssetRoot.
	SudoRun func(opts Options) error

	// ApplicationsDirs overrides dirs scanned for iTerm bak apps (tests).
	// Empty → {Home}/Applications and /Applications.
	ApplicationsDirs []string
	// IsAppRunning injects "is this .app bundle running?" (tests).
	// Nil → ps(1)-based check; on ps failure treat as running (skip delete).
	IsAppRunning func(appPath string) bool

	// Only selects exactly these step IDs (no implied defaults). Include adds
	// aggressive / opt-in step IDs on top of the defaults. Mutually exclusive.
	Only    []string
	Include []string
	// IncludeDirsCleaner purges /private/var/dirs_cleaner (admin) — --dirs-cleaner.
	IncludeDirsCleaner bool

	// DirsCleanerRoot overrides /private/var/dirs_cleaner (tests). A non-default
	// root is never elevated: it must be accessible to this process.
	DirsCleanerRoot string
	// LogShow injects `log` (args after the binary) for purge health (tests).
	LogShow func(args []string) ([]byte, error)
	// TempDirs overrides the temp roots probed for long paths (tests).
	TempDirs []string
	// Elevate injects privileged script execution (tests). Nil → ElevateMode.
	Elevate func(purpose, script string) ([]byte, error)
	// ElevateMode chooses sudo (CLI, needs a TTY) or the osascript dialog (web).
	ElevateMode string

	// OnStep receives progress events (web streaming). Nil → none.
	OnStep func(StepEvent)
	// HistoryFile appends one RunRecord per run when set (CLI/web).
	HistoryFile string
	// Source labels the run record ("cli" or "web").
	Source string
}

// Elevation modes for admin-only steps.
const (
	ElevateSudo      = "sudo"
	ElevateOSAScript = "osascript"
)

// StepEvent is a progress notification for one step.
type StepEvent struct {
	Type   string `json:"type"` // start | done
	ID     string `json:"id"`
	Before int64  `json:"before,omitempty"`
	After  int64  `json:"after,omitempty"`
	Freed  int64  `json:"freed,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Result is the pipeline outcome.
type Result struct {
	Steps      []StepResult `json:"steps"`
	TotalFreed int64        `json:"totalFreed"`
}

// StepResult is one step's size accounting.
type StepResult struct {
	ID     string `json:"id"`
	Before int64  `json:"before"`
	After  int64  `json:"after"`
	Freed  int64  `json:"freed"`
	// Error is the warning text when the step failed (run continued).
	Error string `json:"error,omitempty"`
}

// Run executes the five-step pipeline with injectable Home/writers/PathEnv.
// Dry-run always inventories iOS Simulator runtimes (when xcrun is available)
// and annotates that they are not deleted by default unless IncludeIOSSim.
// Live delete of runtimes only when IncludeIOSSim is true.
func Run(opts Options) (*Result, error) {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}
	if opts.Home == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolve home: %w", err)
		}
		opts.Home = home
	}
	opts, err := applySelection(opts)
	if err != nil {
		return nil, err
	}
	started := time.Now()
	result, runErr := runPipeline(opts)
	recordRun(opts, started, result, runErr)
	return result, runErr
}

func emitStep(opts Options, ev StepEvent) {
	if opts.OnStep != nil {
		opts.OnStep(ev)
	}
}

func runPipeline(opts Options) (*Result, error) {
	steps := selectSteps(opts)
	result := &Result{
		Steps: make([]StepResult, 0, len(steps)+2),
	}

	var defaultTotal, aggressiveTotal int64
	attempted, failed := 0, 0
	addStep := func(sr StepResult) {
		result.Steps = append(result.Steps, sr)
		emitStep(opts, StepEvent{Type: "done", ID: sr.ID, Before: sr.Before, After: sr.After, Freed: sr.Freed, Error: sr.Error})
	}

	for _, def := range steps {
		sr := StepResult{ID: def.ID}
		emitStep(opts, StepEvent{Type: "start", ID: def.ID})
		sr.Before = def.Measure(opts)

		if opts.DryRun {
			// Would free = before; no apply.
			sr.After = sr.Before
			sr.Freed = sr.Before
			addStep(sr)
			result.TotalFreed += sr.Freed
			if def.Tier == TierDefault {
				defaultTotal += sr.Freed
			} else {
				aggressiveTotal += sr.Freed
			}
			continue
		}

		attempted++
		if def.Apply != nil {
			if err := def.Apply(opts); err != nil {
				if isToolMissing(err) {
					warnf(opts.Stderr, "%s not found; skipping %s", stepToolName(def.ID), def.ID)
					// Skip apply: remeasure (after stays full), freed 0.
					sr.After = def.Measure(opts)
					sr.Freed = 0
					addStep(sr)
					continue
				}
				// Non-fatal tool failure: warn, remeasure, continue with the
				// remaining steps. One broken cache never aborts the run.
				warnf(opts.Stderr, "%s failed: %v", def.ID, err)
				sr.Error = err.Error()
				failed++
			}
		}

		sr.After = def.Measure(opts)
		freed := sr.Before - sr.After
		if freed < 0 {
			freed = 0
		}
		sr.Freed = freed
		addStep(sr)
		result.TotalFreed += sr.Freed
		if def.Tier == TierDefault {
			defaultTotal += sr.Freed
		} else {
			aggressiveTotal += sr.Freed
		}
	}

	var inv iosInventory
	// Dry-run always inventories iOS for print-all + annotation.
	// Live measures/applies only when IncludeIOSSim.
	if opts.DryRun || opts.IncludeIOSSim {
		inv = measureIOSSim(opts)
	}

	// MobileAsset size (always on dry-run; live when --ios-sim-assets).
	var assetRoot string
	var assetSize int64
	var assetErr error
	if opts.DryRun || opts.IncludeIOSSimAssets {
		assetRoot, assetSize, assetErr = measureIOSSimAssets(opts)
	}

	if opts.DryRun {
		if opts.IncludeIOSSim {
			// Include iOS runtimes in primary would-free total and Steps.
			sr := StepResult{
				ID:     "ios-sim-runtimes",
				Before: inv.TotalSize,
				After:  inv.TotalSize, // dry-run: no apply
				Freed:  inv.TotalSize,
			}
			if inv.ToolMissing {
				sr.Before, sr.After, sr.Freed = 0, 0, 0
			}
			addStep(sr)
			result.TotalFreed += sr.Freed
		}
		if opts.IncludeIOSSimAssets {
			sr := StepResult{
				ID:     "ios-sim-mobileasset",
				Before: assetSize,
				After:  assetSize,
				Freed:  assetSize,
			}
			if assetErr != nil {
				sr.Before, sr.After, sr.Freed = 0, 0, 0
			}
			addStep(sr)
			result.TotalFreed += sr.Freed
		}
		if opts.IncludeDirsCleaner {
			size, _ := measureDirsCleaner(opts)
			addStep(StepResult{ID: stepDirsCleaner, Before: size, After: size, Freed: size})
			result.TotalFreed += size
		}
		if err := printReport(opts, result, defaultTotal, aggressiveTotal, inv, assetRoot, assetSize, assetErr); err != nil {
			return result, err
		}
		return result, nil
	}

	// Live path: optional iOS runtime apply.
	if opts.IncludeIOSSim {
		attempted++
		sr := StepResult{ID: "ios-sim-runtimes"}
		emitStep(opts, StepEvent{Type: "start", ID: sr.ID})
		if inv.ToolMissing {
			warnf(opts.Stderr, "xcrun not found; skipping ios-sim-runtimes")
			sr.Before, sr.After, sr.Freed = 0, 0, 0
			addStep(sr)
		} else if inv.Err != nil {
			warnf(opts.Stderr, "simctl runtime list failed: %v", inv.Err)
			sr.Error = inv.Err.Error()
			failed++
			sr.Before, sr.After, sr.Freed = 0, 0, 0
			addStep(sr)
		} else if len(inv.Runtimes) == 0 {
			// No runtimes installed: `simctl runtime delete all` exits non-zero
			// here ("No matching images found to delete"), so skip the call
			// rather than reporting a warning for a no-op.
			addStep(sr)
		} else {
			sr.Before = inv.TotalSize
			if err := applyIOSSimDelete(opts); err != nil {
				if isToolMissing(err) {
					warnf(opts.Stderr, "xcrun not found; skipping ios-sim-runtimes")
					sr.After = sr.Before
					sr.Freed = 0
				} else {
					warnf(opts.Stderr, "ios-sim-runtimes failed: %v", err)
					sr.Error = err.Error()
					failed++
					sr.After = remeasureIOSAfterDelete(opts)
					freed := sr.Before - sr.After
					if freed < 0 {
						freed = 0
					}
					sr.Freed = freed
				}
			} else {
				sr.After = remeasureIOSAfterDelete(opts)
				freed := sr.Before - sr.After
				if freed < 0 {
					freed = 0
				}
				sr.Freed = freed
			}
			addStep(sr)
			result.TotalFreed += sr.Freed
		}
	}

	// Live path: optional MobileAsset wipe (may require sudo).
	// A privilege/SIP failure only warns: the remaining steps still run and the
	// report is still printed. Exit stays 0 unless every step failed.
	if opts.IncludeIOSSimAssets {
		attempted++
		sr := StepResult{ID: "ios-sim-mobileasset"}
		emitStep(opts, StepEvent{Type: "start", ID: sr.ID})
		before, after, err := applyIOSSimAssets(opts)
		sr.Before = before
		sr.After = after
		freed := before - after
		if freed < 0 {
			freed = 0
		}
		sr.Freed = freed
		if err != nil {
			warnf(opts.Stderr, "ios-sim-mobileasset failed: %v", err)
			sr.Error = err.Error()
			if assetErr == nil {
				// Surface the SIP recipe in the report as well.
				assetErr = err
			}
			failed++
		}
		addStep(sr)
		result.TotalFreed += sr.Freed
	}

	// Live path: optional dirs_cleaner purge (admin). Failure only warns.
	if opts.IncludeDirsCleaner {
		attempted++
		sr := StepResult{ID: stepDirsCleaner}
		emitStep(opts, StepEvent{Type: "start", ID: sr.ID})
		out, err := PurgeDirsCleaner(opts)
		if err != nil {
			warnf(opts.Stderr, "dirs-cleaner failed: %v", err)
			sr.Error = err.Error()
			failed++
		} else {
			sr.Before, sr.After = out.Before, out.After
			if sr.Freed = out.Before - out.After; sr.Freed < 0 {
				sr.Freed = 0
			}
			if len(out.Restricted) > 0 {
				warnf(opts.Stderr, "dirs-cleaner left %d restricted files (e.g. %s); SIP blocks removal — see `clear-tmp dirs-cleaner how-to`", len(out.Restricted), out.Restricted[0])
			} else if out.Left > 0 {
				warnf(opts.Stderr, "dirs-cleaner left %d entries under %s", out.Left, dirsCleanerRoot(opts))
			}
		}
		addStep(sr)
		result.TotalFreed += sr.Freed
	}

	if err := printReport(opts, result, defaultTotal, aggressiveTotal, inv, assetRoot, assetSize, assetErr); err != nil {
		return result, err
	}

	// Only a run where every selected step failed is a real error.
	if attempted > 0 && failed == attempted {
		return result, fmt.Errorf("all %d selected steps failed (see warnings above)", failed)
	}
	return result, nil
}

// selectSteps filters the registry by each step's Enabled gate.
func selectSteps(opts Options) []stepDef {
	all := allSteps()
	out := make([]stepDef, 0, len(all))
	for _, def := range all {
		if !stepEnabled(opts, def) {
			continue
		}
		out = append(out, def)
	}
	return out
}

// stepToolName maps a step ID to the external tool it shells out to (for
// missing-tool warnings).
func stepToolName(stepID string) string {
	switch stepID {
	case "brew-cleanup":
		return "brew"
	case "npm-cache":
		return "npm"
	case "go-build-cache":
		return "go"
	case "pnpm-store":
		return "pnpm"
	}
	return stepID
}

func printReport(opts Options, result *Result, defaultTotal, aggressiveTotal int64, inv iosInventory, assetRoot string, assetSize int64, assetErr error) error {
	w := opts.Stdout
	use := opts.UseColor

	if opts.DryRun {
		fmt.Fprintln(w, gray(use, "[dry-run]"))
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "Step\tWould free")
		for _, s := range result.Steps {
			label := s.ID
			if s.ID == "ios-sim-runtimes" && opts.IncludeIOSSim {
				label = s.ID + " (would delete)"
			}
			if s.ID == "ios-sim-mobileasset" && opts.IncludeIOSSimAssets {
				label = s.ID + " (would remove)"
			}
			if s.ID == stepMSPlaywright && (opts.Aggressive || opts.IncludeMSPlaywright) {
				label = s.ID + " (would delete)"
			}
			if s.ID == stepOpenCodeSnapshot && (opts.Aggressive || opts.IncludeOpenCodeSnapshot) {
				label = s.ID + " (would delete)"
			}
			if s.ID == stepDirsCleaner {
				label = s.ID + " (would purge)"
				if _, known := measureDirsCleaner(opts); !known {
					label = s.ID + " (would purge; size needs admin)"
				}
			}
			fmt.Fprintf(tw, "%s\t%s\n", label, humanSize(s.Freed))
		}
		_ = tw.Flush()

		// Always print iOS inventory + annotation on dry-run.
		printIOSInventory(w, use, inv, opts.IncludeIOSSim)
		fmt.Fprintln(w, "note: iOS Simulator runtimes are not deleted by default; pass --ios-sim to include.")

		// MobileAsset section (print-all).
		if assetErr != nil {
			fmt.Fprintln(w, gray(use, "iOS Simulator MobileAsset:"))
			fmt.Fprintf(w, "  (error: %v)\n", assetErr)
		} else if assetRoot != "" {
			printIOSSimAssetsSection(w, use, assetRoot, assetSize, opts.IncludeIOSSimAssets)
		}
		fmt.Fprintln(w, "note: MobileAsset needs --ios-sim-assets; if SIP blocks, use Recovery csrutil recipe above.")

		pwSize := measureMSPlaywright(opts)
		ocSnapSize := measureOpenCodeSnapshot(opts)
		printOptInCacheSection(w, use, "Playwright browsers (Library/Caches/ms-playwright)",
			msPlaywrightPath(opts.Home), pwSize, opts.Aggressive || opts.IncludeMSPlaywright, "--ms-playwright")
		printOptInCacheSection(w, use, "OpenCode snapshots (.local/share/opencode/snapshot)",
			openCodeSnapshotPath(opts.Home), ocSnapSize, opts.Aggressive || opts.IncludeOpenCodeSnapshot, "--opencode-snapshot")

		// iTerm bak apps: default step; print detailed inventory on dry-run.
		printITermBakSection(w, use, listITermBakApps(opts), true)

		// ShipIt updater caches: inventory only; never deleted; not in TotalFreed.
		printShipItSection(w, use, measureShipIt(opts.Home))

		// dirs_cleaner: metadata only (no root, no log query) on every dry-run.
		printDirsCleanerSection(w, use, DirsCleanerBasic(opts), opts.IncludeDirsCleaner)

		withIOS := defaultTotal + inv.TotalSize
		withAssets := defaultTotal + assetSize
		withAll := defaultTotal + inv.TotalSize + assetSize
		// --aggressive-all implies --ios-sim and --ios-sim-assets.
		withAggressive := defaultTotal + aggressiveTotal + inv.TotalSize + assetSize

		defaultLine := fmt.Sprintf("Total would free (default steps): %s (%d B)", humanSize(defaultTotal), defaultTotal)
		withIOSLine := fmt.Sprintf("Total would free (with --ios-sim): %s (%d B)", humanSize(withIOS), withIOS)
		withAssetsLine := fmt.Sprintf("Total would free (with --ios-sim-assets): %s (%d B)", humanSize(withAssets), withAssets)
		withAllLine := fmt.Sprintf("Total would free (with --ios-sim and --ios-sim-assets): %s (%d B)", humanSize(withAll), withAll)
		withAggressiveLine := fmt.Sprintf("Total would free (with --aggressive-all): %s (%d B)", humanSize(withAggressive), withAggressive)

		anyOptIn := opts.Aggressive || opts.IncludeIOSSim || opts.IncludeIOSSimAssets ||
			opts.IncludeMSPlaywright || opts.IncludeOpenCodeSnapshot || opts.IncludeDirsCleaner ||
			len(opts.Only) > 0 || len(opts.Include) > 0
		if anyOptIn {
			primary := fmt.Sprintf("Total would free: %s (%d B)", humanSize(result.TotalFreed), result.TotalFreed)
			if result.TotalFreed > 0 {
				primary = green(use, primary)
			}
			fmt.Fprintln(w, primary)
		} else if defaultTotal > 0 {
			defaultLine = green(use, defaultLine)
		}
		fmt.Fprintln(w, defaultLine)
		if opts.Aggressive {
			// --aggressive-all implies the iOS opt-ins, so the per-flag lines
			// would all be lower than the aggressive total: show one line.
			fmt.Fprintln(w, withAggressiveLine)
			if assetSize > 0 {
				effective := withAggressive - assetSize
				fmt.Fprintf(w, "note: --aggressive-all includes %s of MobileAsset that SIP usually blocks;\n", humanSize(assetSize))
				fmt.Fprintf(w, "      effective aggressive total without it: %s (%d B)\n", humanSize(effective), effective)
			}
		} else {
			fmt.Fprintln(w, withIOSLine)
			fmt.Fprintln(w, withAssetsLine)
			fmt.Fprintln(w, withAllLine)
			if opts.IncludeMSPlaywright {
				fmt.Fprintf(w, "Total would free (with --ms-playwright): %s (%d B)\n", humanSize(defaultTotal+pwSize), defaultTotal+pwSize)
			}
			if opts.IncludeOpenCodeSnapshot {
				fmt.Fprintf(w, "Total would free (with --opencode-snapshot): %s (%d B)\n", humanSize(defaultTotal+ocSnapSize), defaultTotal+ocSnapSize)
			}
		}
	} else {
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "Step\tBefore\tAfter\tFreed")
		for _, s := range result.Steps {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
				s.ID, humanSize(s.Before), humanSize(s.After), humanSize(s.Freed))
		}
		_ = tw.Flush()
		// If live --ios-sim left assets behind, hint.
		if opts.IncludeIOSSim && !opts.IncludeIOSSimAssets && assetRoot == "" {
			// measure assets for hint only
			if r, sz, err := measureIOSSimAssets(opts); err == nil && sz > 0 {
				warnf(opts.Stderr, "iOS Simulator MobileAsset cache still %s at %s; pass --ios-sim-assets (sudo may prompt) to remove", humanSize(sz), r)
			}
		}
		totalLine := fmt.Sprintf("Total freed: %s (%d B)", humanSize(result.TotalFreed), result.TotalFreed)
		if result.TotalFreed > 0 {
			totalLine = green(use, totalLine)
		}
		fmt.Fprintln(w, totalLine)
	}

	if !opts.SkipVolume {
		// Best-effort volume free delta on Home's filesystem.
		before, after, ok := volumeFreeDelta(opts.Home)
		if ok {
			delta := after - before
			line := fmt.Sprintf("Volume free: before %s after %s delta %s",
				humanSize(before), humanSize(after), humanSize(delta))
			fmt.Fprintln(w, gray(use, line))
		}
	}

	return nil
}

func printIOSInventory(w io.Writer, use bool, inv iosInventory, included bool) {
	fmt.Fprintln(w, gray(use, "iOS Simulator runtimes:"))
	if inv.ToolMissing {
		fmt.Fprintln(w, "  (xcrun not found; inventory unavailable)")
		return
	}
	if inv.Err != nil {
		fmt.Fprintf(w, "  (inventory error: %v)\n", inv.Err)
		return
	}
	if len(inv.Runtimes) == 0 {
		fmt.Fprintln(w, "  (none)")
		return
	}
	for _, r := range inv.Runtimes {
		status := "not deleted by default"
		if included {
			status = "would delete / included"
		}
		fmt.Fprintf(w, "  %s  %s (%d B)  [%s]\n", r.Name, humanSize(r.SizeBytes), r.SizeBytes, status)
	}
	fmt.Fprintf(w, "  iOS Simulator runtimes total: %s (%d B)\n", humanSize(inv.TotalSize), inv.TotalSize)
}

// volumeFreeDelta is implemented per-OS; stub returns ok=false when hard.
func volumeFreeDelta(home string) (before, after int64, ok bool) {
	// Optional secondary metric; tests set SkipVolume=true.
	// Without SkipVolume we still skip if we cannot measure reliably.
	_ = home
	return 0, 0, false
}
