package reclaim

import "io"

// PathMax is the longest path the next boot's dirs_cleaner purge handles
// without stalling.
const PathMax = pathMax

// DirsCleanerRoot returns the dirs_cleaner root that opts targets.
func DirsCleanerRoot(opts Options) string { return dirsCleanerRoot(opts) }

// TruncMiddle shortens s to n runes, keeping both ends.
func TruncMiddle(s string, n int) string { return truncMiddle(s, n) }

// PrintDirsCleanerSection prints the dirs_cleaner status footer used by the
// clear-tmp CLI. It lives here so every consumer renders the same text.
func PrintDirsCleanerSection(w io.Writer, useColor bool, st DirsCleanerStatus, included bool) {
	printDirsCleanerSection(w, useColor, st, included)
}

// Opt-in step IDs, exposed so a shell can wire its own flags to them.
const (
	StepIOSSimRuntimes    = stepIOSSimRuntimes
	StepIOSSimMobileAsset = stepIOSSimMobileAsst
	StepMSPlaywright      = stepMSPlaywright
	StepOpenCodeSnapshot  = stepOpenCodeSnapshot
	StepDirsCleaner       = stepDirsCleaner
)

// ApplySelection validates --only/--include and folds them into opts.
func ApplySelection(opts Options) (Options, error) { return applySelection(opts) }

// RunExternalTool runs tool with args, honouring opts.PathEnv and silencing
// stdout. It returns a *ToolMissingError when the tool is not installed.
func RunExternalTool(opts Options, tool string, args ...string) error {
	return runExternalTool(opts, tool, args...)
}

// RunExternalToolOutput is RunExternalTool with stdout captured.
func RunExternalToolOutput(opts Options, tool string, args ...string) ([]byte, error) {
	return runExternalToolOutput(opts, tool, args...)
}

// IsToolMissing reports whether err means "tool not installed".
func IsToolMissing(err error) bool { return isToolMissing(err) }

// LongPathWarn is the path length where a dirs_cleaner purge starts to stall.
const LongPathWarn = longPathWarn

// DefaultStepIDs lists the step IDs a plain run selects.
func DefaultStepIDs() []string { return defaultStepIDs() }

// ValidateStepIDs checks --only/--include values.
func ValidateStepIDs(flag string, ids []string) error { return validateStepIDs(flag, ids) }

// LogShowArgs are the `log show` arguments used to read dirs_cleaner history.
func LogShowArgs(since string) []string { return logShowArgs(since) }

// ShQuote single-quotes s for a shell.
func ShQuote(s string) string { return shQuote(s) }
