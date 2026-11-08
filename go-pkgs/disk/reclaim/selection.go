package reclaim

import (
	"fmt"
	"sort"
	"strings"
)

// Step IDs that live outside the stepDef registry (handled specially in Run).
const (
	stepIOSSimRuntimes   = "ios-sim-runtimes"
	stepIOSSimMobileAsst = "ios-sim-mobileasset"
	stepDirsCleaner      = "dirs-cleaner"
)

// knownStepIDs lists every selectable step ID in execution order.
func knownStepIDs() []string {
	ids := make([]string, 0, 24)
	for _, def := range allSteps() {
		ids = append(ids, def.ID)
	}
	return append(ids, stepIOSSimRuntimes, stepIOSSimMobileAsst, stepDirsCleaner)
}

// defaultStepIDs lists the TierDefault registry IDs (always run without --only).
func defaultStepIDs() []string {
	var ids []string
	for _, def := range defaultSteps() {
		ids = append(ids, def.ID)
	}
	return ids
}

// splitIDList flattens repeated and comma-separated --only / --include values.
func splitIDList(values []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, v := range values {
		for _, p := range strings.Split(v, ",") {
			p = strings.TrimSpace(p)
			if p == "" || seen[p] {
				continue
			}
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// validateStepIDs reports the first unknown ID for flag.
func validateStepIDs(flag string, ids []string) error {
	known := knownStepIDs()
	for _, id := range ids {
		if !containsString(known, id) {
			return fmt.Errorf("%s: unknown step %q (known: %s)", flag, id, strings.Join(known, ", "))
		}
	}
	return nil
}

// applySelection folds Only / Include into the per-step Include* flags so the
// rest of Run keeps working off the original booleans.
//
// Only means "exactly these steps": default steps are not implied and
// --aggressive-all is off. Include adds aggressive / opt-in steps on top of the
// defaults.
func applySelection(opts Options) (Options, error) {
	opts.Only = splitIDList(opts.Only)
	opts.Include = splitIDList(opts.Include)
	if len(opts.Only) > 0 && len(opts.Include) > 0 {
		return opts, fmt.Errorf("--only and --include are mutually exclusive")
	}
	if err := validateStepIDs("--only", opts.Only); err != nil {
		return opts, err
	}
	if err := validateStepIDs("--include", opts.Include); err != nil {
		return opts, err
	}
	if len(opts.Only) > 0 {
		opts.Aggressive = false
		opts.IncludeIOSSim = containsString(opts.Only, stepIOSSimRuntimes)
		opts.IncludeIOSSimAssets = containsString(opts.Only, stepIOSSimMobileAsst)
		opts.IncludeMSPlaywright = containsString(opts.Only, stepMSPlaywright)
		opts.IncludeOpenCodeSnapshot = containsString(opts.Only, stepOpenCodeSnapshot)
		opts.IncludeDirsCleaner = containsString(opts.Only, stepDirsCleaner)
		return opts, nil
	}
	inc := opts.Include
	opts.IncludeIOSSim = opts.IncludeIOSSim || containsString(inc, stepIOSSimRuntimes)
	opts.IncludeIOSSimAssets = opts.IncludeIOSSimAssets || containsString(inc, stepIOSSimMobileAsst)
	opts.IncludeMSPlaywright = opts.IncludeMSPlaywright || containsString(inc, stepMSPlaywright)
	opts.IncludeOpenCodeSnapshot = opts.IncludeOpenCodeSnapshot || containsString(inc, stepOpenCodeSnapshot)
	opts.IncludeDirsCleaner = opts.IncludeDirsCleaner || containsString(inc, stepDirsCleaner)
	return opts, nil
}

// stepEnabled is the registry gate after applySelection.
func stepEnabled(opts Options, def stepDef) bool {
	if len(opts.Only) > 0 {
		return containsString(opts.Only, def.ID)
	}
	if def.Enabled == nil {
		return true
	}
	return def.Enabled(opts) || containsString(opts.Include, def.ID)
}

// selectedStepIDs returns every step ID the (selection-applied) opts would run.
func selectedStepIDs(opts Options) []string {
	var ids []string
	for _, def := range allSteps() {
		if stepEnabled(opts, def) {
			ids = append(ids, def.ID)
		}
	}
	if opts.IncludeIOSSim {
		ids = append(ids, stepIOSSimRuntimes)
	}
	if opts.IncludeIOSSimAssets {
		ids = append(ids, stepIOSSimMobileAsst)
	}
	if opts.IncludeDirsCleaner {
		ids = append(ids, stepDirsCleaner)
	}
	return ids
}

// CommandFor renders the canonical CLI invocation for an exact step selection:
// all defaults present → `clear-tmp [--include extras]`; otherwise `--only`.
// Extras/only follow registry order so the same selection prints the same line.
func CommandFor(selected []string, dryRun bool) string {
	order := map[string]int{}
	for i, id := range knownStepIDs() {
		order[id] = i
	}
	sel := append([]string(nil), selected...)
	sort.SliceStable(sel, func(i, j int) bool { return order[sel[i]] < order[sel[j]] })

	parts := []string{"clear-tmp"}
	if dryRun {
		parts = append(parts, "--dry-run")
	}
	defaults := defaultStepIDs()
	allDefaults := true
	for _, id := range defaults {
		if !containsString(sel, id) {
			allDefaults = false
			break
		}
	}
	if !allDefaults {
		return strings.Join(append(parts, "--only", strings.Join(sel, ",")), " ")
	}
	var extra []string
	for _, id := range sel {
		if !containsString(defaults, id) {
			extra = append(extra, id)
		}
	}
	if len(extra) > 0 {
		parts = append(parts, "--include", strings.Join(extra, ","))
	}
	return strings.Join(parts, " ")
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
