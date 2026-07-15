package tagscope

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

// ApplyOptions controls tag creation and push behavior.
type ApplyOptions struct {
	DryRun bool
	Push   bool
}

// ApplyResult reports tags created or skipped during Apply.
type ApplyResult struct {
	Created []string // tag names created
	Skipped []ScopeDecision
	Pushed  []string
}

const humanColTag = 13

// Plan runs Collect + LoadOwnedTrees + Evaluate for repoRoot at headRef.
func Plan(repoRoot string, headRef string) (ChangePlan, CollectedTags, error) {
	collected, err := Collect(repoRoot)
	if err != nil {
		return ChangePlan{}, CollectedTags{}, err
	}

	headCommit, err := resolveCommit(repoRoot, headRef)
	if err != nil {
		return ChangePlan{}, CollectedTags{}, err
	}

	ownedTrees, err := LoadOwnedTrees(repoRoot, collected, headRef)
	if err != nil {
		return ChangePlan{}, CollectedTags{}, err
	}

	releaseCommits, err := releaseCommitsForScopes(repoRoot, collected)
	if err != nil {
		return ChangePlan{}, CollectedTags{}, err
	}

	plan := Evaluate(ChangeCheckInput{
		Collected:      collected,
		HeadCommit:     headCommit,
		ReleaseCommits: releaseCommits,
		OwnedTrees:     ownedTrees,
	})
	return plan, collected, nil
}

func releaseCommitsForScopes(repoRoot string, collected CollectedTags) (map[TagScopeKey]string, error) {
	out := make(map[TagScopeKey]string, len(collected.Scopes))
	for _, scope := range collected.Scopes {
		key := scopeKey(scope)
		lineage := collected.ByScope[key]
		if lineage.LatestRelease == nil {
			continue
		}
		commit, err := resolveCommit(repoRoot, lineage.LatestRelease.FullName)
		if err != nil {
			return nil, fmt.Errorf("resolve release %q: %w", lineage.LatestRelease.FullName, err)
		}
		out[key] = commit
	}
	return out, nil
}

// Apply creates lightweight tags at headRef for planned decisions unless DryRun.
// When Push is set, each created tag is pushed to origin (tag ref only).
func Apply(repoRoot string, plan ChangePlan, headRef string, opts ApplyOptions) (ApplyResult, error) {
	headCommit, err := resolveCommit(repoRoot, headRef)
	if err != nil {
		return ApplyResult{}, err
	}

	result := ApplyResult{
		Skipped: make([]ScopeDecision, 0),
		Created: make([]string, 0),
		Pushed:  make([]string, 0),
	}

	for _, decision := range plan.Decisions {
		if decision.NextTag == "" {
			result.Skipped = append(result.Skipped, decision)
			continue
		}
		if opts.DryRun {
			continue
		}
		if err := createLightweightTag(repoRoot, decision.NextTag, headCommit); err != nil {
			return result, err
		}
		result.Created = append(result.Created, decision.NextTag)
		if opts.Push {
			if err := pushTag(repoRoot, decision.NextTag); err != nil {
				return result, err
			}
			result.Pushed = append(result.Pushed, decision.NextTag)
		}
	}
	return result, nil
}

func createLightweightTag(repoRoot, name, commit string) error {
	cmd := exec.Command("git", "tag", name, commit)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git tag %s %s in %s: %w\n%s", name, commit, repoRoot, err, out)
	}
	return nil
}

func pushTag(repoRoot, name string) error {
	cmd := exec.Command("git", "push", "origin", name)
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git push origin %s in %s: %w\n%s", name, repoRoot, err, out)
	}
	return nil
}

// FormatPlanHuman renders per-scope decision lines for stdout (no summary footer).
func FormatPlanHuman(plan ChangePlan, collected CollectedTags) string {
	var b strings.Builder
	for _, decision := range plan.Decisions {
		b.WriteString(formatDecisionLine(decision, collected))
		b.WriteByte('\n')
	}
	return b.String()
}

func formatDecisionLine(decision ScopeDecision, collected CollectedTags) string {
	tagCol := padRight(decision.LatestRelease, humanColTag)
	reason := humanReasonForDecision(decision, collected)
	reasonCol := padRight(" "+reason, humanReasonWidth(reason))
	if decision.NextTag != "" {
		return tagCol + reasonCol + " ->  " + decision.NextTag
	}
	return tagCol + reasonCol + " ->  skip"
}

func humanReasonWidth(reason string) int {
	// Leading separator space is included in the padded reason field.
	if len(reason) >= 30 {
		return 32
	}
	return 31
}

func humanReasonForDecision(decision ScopeDecision, collected CollectedTags) string {
	switch decision.SkipReason {
	case "no-changes":
		return "no changes"
	case "same-commit":
		return "same commit"
	case "no-baseline":
		return "no baseline"
	case "prerelease-head":
		key := scopeKey(decision.Scope)
		lineage := collected.ByScope[key]
		if lineage.Newest != nil {
			return "prerelease head (" + lineage.Newest.FullName + ")"
		}
		return "prerelease head"
	case "":
		return "owned changed"
	default:
		return decision.SkipReason
	}
}

func padRight(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// PlannedTagCount returns how many scopes plan a next tag.
func PlannedTagCount(plan ChangePlan) int {
	n := 0
	for _, d := range plan.Decisions {
		if d.NextTag != "" {
			n++
		}
	}
	return n
}

// FormatPlanSummary returns the human footer line for dry-run or apply.
func FormatPlanSummary(plan ChangePlan, dryRun bool) string {
	n := PlannedTagCount(plan)
	word := "created"
	if dryRun {
		word = "planned"
	}
	return fmt.Sprintf("%d tag %s", n, word)
}

// FormatTaggedLines returns apply-mode tagged lines (one per created tag).
func FormatTaggedLines(repoRoot string, headRef string, created []string) (string, error) {
	if len(created) == 0 {
		return "", nil
	}
	short, err := shortCommit(repoRoot, headRef)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, name := range created {
		fmt.Fprintf(&b, "tagged %s @ %s\n", name, short)
	}
	return b.String(), nil
}

func shortCommit(repoRoot, ref string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short=7", ref)
	cmd.Dir = repoRoot
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse --short=7 %s in %s: %w", ref, repoRoot, err)
	}
	return strings.TrimSpace(string(out)), nil
}

type planJSONDecision struct {
	Scope      TagScope `json:"scope"`
	LatestTag  string   `json:"latest_tag"`
	NextTag    string   `json:"next_tag,omitempty"`
	SkipReason string   `json:"skip_reason,omitempty"`
}

type planJSONSummary struct {
	Planned int `json:"planned"`
	Created int `json:"created,omitempty"`
}

type planJSON struct {
	Head      string             `json:"head"`
	Decisions []planJSONDecision `json:"decisions"`
	Summary   planJSONSummary    `json:"summary"`
}

// FormatPlanJSON renders machine-readable plan/result for --json.
func FormatPlanJSON(plan ChangePlan, collected CollectedTags, dryRun bool, created int) (string, error) {
	decisions := make([]planJSONDecision, 0, len(plan.Decisions))
	for _, d := range plan.Decisions {
		decisions = append(decisions, planJSONDecision{
			Scope:      d.Scope,
			LatestTag:  d.LatestRelease,
			NextTag:    d.NextTag,
			SkipReason: d.SkipReason,
		})
	}
	payload := planJSON{
		Head:      plan.Head,
		Decisions: decisions,
		Summary: planJSONSummary{
			Planned: PlannedTagCount(plan),
		},
	}
	if !dryRun {
		payload.Summary.Created = created
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}