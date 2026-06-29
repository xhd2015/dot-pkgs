package worktree

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// FormatPlanPrompt returns the confirmation prompt for a merge-back plan,
// including the concrete git commands that would run.
func FormatPlanPrompt(plan MergeBackPlan) string {
	var b strings.Builder
	// Doctest output templates begin with a newline after the opening backtick,
	// which parses as a leading empty literal line in assert.Output.
	b.WriteByte('\n')
	switch plan.Relation {
	case "ahead":
		fmt.Fprintf(&b, "branch %s is ahead, merge into %s?\n", plan.Branch, plan.TargetLabel)
	case "diverged":
		fmt.Fprintf(&b, "branch %s has diverged, rebase and merge into %s?\n", plan.Branch, plan.TargetLabel)
	default:
		fmt.Fprintf(&b, "proceed with merge-back for branch %s?\n", plan.Branch)
	}
	WritePlannedCommandsDisplay(&b, plan)
	b.WriteString("Proceed? [Y/n]:")
	return b.String()
}

// PromptConfirmPlan prompts the user to confirm a merge-back plan.
// When confirmFromStdin is true, confirmation is read from piped stdin.
func PromptConfirmPlan(plan MergeBackPlan, confirmFromStdin bool) (bool, error) {
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	if !isTTY {
		if stdinIsPipe() {
			if !confirmFromStdin {
				return false, fmt.Errorf("stdin is not a terminal; pass --confirm-from-stdin to read confirmation from piped stdin")
			}
		} else {
			return false, ErrConfirmationRequired
		}
	}

	prompt := FormatPlanPrompt(plan)
	if isTTY {
		fmt.Fprint(os.Stderr, prompt+" ")
	} else {
		fmt.Fprint(os.Stdout, prompt+" ")
	}

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("read input: %w", err)
	}
	line = strings.TrimSpace(strings.ToLower(line))
	switch line {
	case "", "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid input: %q (expected y/n)", strings.TrimSpace(line))
	}
}

func stdinIsPipe() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}